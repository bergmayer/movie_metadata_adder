#!/bin/bash
set -e

# Hard-code the TMDB API key
tmdb_api_key="a8b8f7233221f1a507ccedf34c2324f8"

if [[ $# -ne 1 ]]; then
    echo "Usage: $0 movie_file.mkv"
    exit 1
fi

movie_file="$1"

# URL encode function
urlencode() {
    local string="${1}"
    local strlen=${#string}
    local encoded=""
    local pos c o

    for (( pos=0 ; pos<strlen ; pos++ )); do
        c=${string:$pos:1}
        case "$c" in
            [-_.~a-zA-Z0-9] ) o="${c}" ;;
            * )               printf -v o '%%%02x' "'$c"
        esac
        encoded+="${o}"
    done
    echo "${encoded}"
}

# Extract probable movie title and year from filename
filename=$(basename "$movie_file")
filename=${filename%.*}
clean_name=$(echo "$filename" | sed 's/[._]/ /g')
year=$(echo "$clean_name" | grep -o '\(?[12][0-9][0-9][0-9]\)?' | head -1 | tr -d '()')
clean_name=$(echo "$clean_name" | sed -E 's/\(?[12][0-9][0-9][0-9]\)?//g' | 
    sed -E 's/720p|1080p|2160p|BRRip|BDRip|BluRay|WEBRip|HDTV|x264|x265|10bit|HEVC|AAC|AC3|DTS|GalaxyRG|RARBG|.*MB.*//g' |
    sed 's/  */ /g' | sed 's/^ *//g' | sed 's/ *$//g')

# Search TMDB API and present choices
declare -a titles=()
declare -a years=()
declare -a ids=()
declare -a overviews=()

for term in "$clean_name" "$(echo "$clean_name" | cut -d' ' -f1-2)" "$(echo "$clean_name" | sed 's/\s.*$//')"; do
    search_url="https://api.themoviedb.org/3/search/movie?api_key=${tmdb_api_key}&query=$(urlencode "$term")"
    [[ -n "$year" ]] && search_url="${search_url}&year=${year}"
    
    search_results=$(curl -s "$search_url")
    
    while IFS= read -r result; do
        [ -z "$result" ] && continue
        
        title=$(echo "$result" | jq -r '.title')
        result_year=$(echo "$result" | jq -r '.release_date' | cut -d"-" -f1)
        id=$(echo "$result" | jq -r '.id')
        overview=$(echo "$result" | jq -r '.overview' | cut -c1-100)
        
        # Check if this ID is already in our array
        if [[ ! " ${ids[@]} " =~ " ${id} " ]]; then
            titles+=("$title")
            years+=("$result_year")
            ids+=("$id")
            overviews+=("$overview")
        fi
    done < <(echo "$search_results" | jq -c '.results[]')
done

if [ ${#ids[@]} -eq 0 ]; then
    echo "No matches found for: $clean_name"
    exit 1
fi

echo "Found ${#ids[@]} potential matches for: $clean_name"
echo "----------------------------------------"
for i in "${!ids[@]}"; do
    echo "[$i] ${titles[$i]} (${years[$i]})"
    echo "    ${overviews[$i]}..."
    echo
done
echo "[q] Quit"
echo "----------------------------------------"

while true; do
    read -p "Select number or 'q' to quit: " choice
    if [[ $choice == "q" ]]; then
        echo "Exiting..."
        exit 0
    elif [[ $choice =~ ^[0-9]+$ ]] && [ "$choice" -lt "${#ids[@]}" ]; then
        tmdb_id=${ids[$choice]}
        break
    else
        echo "Invalid choice. Please try again."
    fi
done

# Fetch movie data
metadata=$(curl -s "https://api.themoviedb.org/3/movie/${tmdb_id}?api_key=${tmdb_api_key}&append_to_response=credits")

# Parse JSON response
title=$(echo "$metadata" | jq -r '.title')
year=$(echo "$metadata" | jq -r '.release_date' | cut -d "-" -f1)
director=$(echo "$metadata" | jq -r '.credits.crew[] | select(.job == "Director") | .name')
actors=$(echo "$metadata" | jq -r '[.credits.cast[] | .name] | join(", ")')
genres=$(echo "$metadata" | jq -r '[.genres[] | .name] | join(", ")')

# Fetch poster
poster_path=$(echo "$metadata" | jq -r '.poster_path')
if [[ "$poster_path" != "null" ]]; then
    poster_url="https://image.tmdb.org/t/p/original${poster_path}"
    curl -s -o "poster.jpg" "$poster_url"
fi

# Add metadata
if [[ -f "poster.jpg" ]]; then
    ffmpeg -i "${movie_file}" -i "poster.jpg" -map 0 -map 1 -metadata title="$title" -metadata year="$year" -metadata director="$director" -metadata actors="$actors" -metadata genre="$genres" -c copy -c:v:1 png -disposition:v:1 attached_pic "temp_${filename}.${movie_file##*.}"
    rm "poster.jpg"
else
    ffmpeg -i "${movie_file}" -metadata title="$title" -metadata year="$year" -metadata director="$director" -metadata actors="$actors" -metadata genre="$genres" -codec copy "temp_${filename}.${movie_file##*.}"
fi

mv "temp_${filename}.${movie_file##*.}" "${title} (${year}).${movie_file##*.}"
echo "Completed: ${title} (${year}).${movie_file##*.}"