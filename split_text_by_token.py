#!/usr/bin/env python3
"""
Split text by tokenizer into JSON array.

token json format:
[
    {
        "content": "hello ",
        "tokens": 1
    },
    {
        "content": "world",
        "tokens": 1
    }
]
"""

import argparse
import json
import sys
import tiktoken


def split_text_by_token(text, encoding_name="cl100k_base"):
    """
    Split text by tokens and merge incomplete multi-byte sequences (fullrune).
    
    Args:
        text: Input text string
        encoding_name: Tokenizer encoding name (default: cl100k_base for GPT-4)
    
    Returns:
        List of dicts with 'content' and 'tokens' keys
    """
    enc = tiktoken.get_encoding(encoding_name)
    result = []
    
    # Encode the entire text into tokens
    token_ids = enc.encode(text)
    
    # Process tokens and merge incomplete UTF-8 sequences
    i = 0
    while i < len(token_ids):
        token_bytes = enc.decode_single_token_bytes(token_ids[i])
        token_count = 1
        
        # Try to decode as UTF-8
        try:
            content = token_bytes.decode('utf-8')
        except UnicodeDecodeError:
            # Incomplete multi-byte sequence - merge with next tokens
            # Keep accumulating bytes until we get a valid UTF-8 sequence
            j = i + 1
            while j < len(token_ids):
                next_token_bytes = enc.decode_single_token_bytes(token_ids[j])
                token_bytes += next_token_bytes
                try:
                    content = token_bytes.decode('utf-8')
                    token_count = j - i + 1
                    i = j  # Update i to skip the merged tokens
                    break
                except UnicodeDecodeError:
                    j += 1
            else:
                # Still can't decode - use replacement character
                content = token_bytes.decode('utf-8', errors='replace')
                token_count = j - i
                i = j - 1
        
        result.append({
            "content": content,
            "tokens": token_count
        })
        
        i += 1
    
    return result


def main():
    parser = argparse.ArgumentParser(
        description="Split text by tokenizer into JSON array with token counts"
    )
    parser.add_argument(
        "input_file",
        help="Input text file path"
    )
    parser.add_argument(
        "-o", "--output",
        help="Output JSON file path (default: input_file.tokens.json)"
    )
    parser.add_argument(
        "-e", "--encoding",
        default="cl100k_base",
        help="Tokenizer encoding name (default: cl100k_base)"
    )
    parser.add_argument(
        "--pretty",
        action="store_true",
        help="Pretty print JSON with indentation"
    )
    
    args = parser.parse_args()
    
    # Read input file
    try:
        with open(args.input_file, 'r', encoding='utf-8') as f:
            text = f.read()
    except FileNotFoundError:
        print(f"Error: Input file '{args.input_file}' not found", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error reading input file: {e}", file=sys.stderr)
        sys.exit(1)
    
    # Process text
    print(f"Processing text with {args.encoding} tokenizer...")
    result = split_text_by_token(text, args.encoding)
    
    # Determine output file
    output_file = args.output or f"{args.input_file}.tokens.json"
    
    # Write output
    try:
        with open(output_file, 'w', encoding='utf-8') as f:
            if args.pretty:
                json.dump(result, f, ensure_ascii=False, indent=2)
            else:
                json.dump(result, f, ensure_ascii=False)
        
        total_chars = len(result)
        total_tokens = sum(item['tokens'] for item in result)
        print(f"Success! Written {total_chars} characters ({total_tokens} tokens) to {output_file}")
        
    except Exception as e:
        print(f"Error writing output file: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()