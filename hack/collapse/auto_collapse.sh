#!/bin/bash

# allow > to overwrite files
set +o noclobber

# Auto Collapse
#This is a script that takes in
# - a markdown file
# - a designated output file
# - the number of lines as a threshold on when to collapse a section.
#
# Sample
# ./auto_collapse.sh test_example.md test_example_collapsible.md 5

# Input and output files
INPUT_FILE=${1:-"README.md"}
OUTPUT_FILE=${2:-"README_collapsible.md"}
THRESHOLD=${3:-10}

# Ensure output file is empty initially
true > "$OUTPUT_FILE"

# Variables to track sections
inside_section=false
section_lines=()
section_header=""
last_line=""

# Function to count changes (lines starting with '*')
count_changes() {
    local lines=("$@")
    local count=0
    for line in "${lines[@]}"; do
        if [[ $line =~ ^\* ]]; then
            ((count++))
        fi
    done
    echo "$count"
}

# Function to process and write a section
write_section() {
    local header="$1"
    local lines=("${@:2}")
    local num_changes
    num_changes=$(count_changes "${lines[@]}")

    # Write the section header as is
    echo "$header" >> "$OUTPUT_FILE"

    if [[ $num_changes -gt $THRESHOLD ]]; then
        # Collapse only the content with a dynamic summary
        {
            echo "<details>"
            echo "<summary>${num_changes} changes</summary>"
            echo ""
            printf "%s\n" "${lines[@]}"
            echo "</details>"
            echo ""
        } >> "$OUTPUT_FILE"
    else
        # Write the content as is if it's below the threshold
        printf "%s\n" "${lines[@]}" >> "$OUTPUT_FILE"
    fi
}

# Read the Markdown file line by line
while IFS= read -r line || [[ -n $line ]]; do
    # Check if the line contains "Full Changelog"
    if [[ $line =~ ^\*\*Full\ Changelog ]]; then
        # Finalize any pending section
        if [[ $inside_section == true ]]; then
            write_section "$section_header" "${section_lines[@]:1}" # Exclude the header
            inside_section=false
        fi
        last_line="$line"
        continue
    fi

    # Preserve comment blocks
    if [[ $line =~ ^\<!-- ]] || [[ $line =~ ^--\> ]]; then
        if [[ $inside_section == true ]]; then
            write_section "$section_header" "${section_lines[@]:1}" # Exclude the header
            inside_section=false
        fi
        echo "$line" >> "$OUTPUT_FILE"
        continue
    fi

    if [[ $line =~ ^#+\  ]]; then # New section starts
        if [[ $inside_section == true ]]; then
            # Write the previous section
            write_section "$section_header" "${section_lines[@]:1}" # Exclude the header
        fi
        # Start a new section
        section_header="$line"
        section_lines=("$line") # Initialize section with the header
        inside_section=true
    else
        # Collect lines of the current section
        section_lines+=("$line")
    fi
done < "$INPUT_FILE"

# Process the last section
if [[ $inside_section == true ]]; then
    write_section "$section_header" "${section_lines[@]:1}" # Exclude the header
fi

# Write the "Full Changelog" line if it exists
if [[ -n $last_line ]]; then
    printf "\n%s" "$last_line" >> "$OUTPUT_FILE"
fi

echo "Collapsible Markdown written to $OUTPUT_FILE"
