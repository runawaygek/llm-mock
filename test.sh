curl --request POST \
  --url http://localhost:8081/v1/chat/completions \
  --header 'Authorization: Bearer sk-wbmgnbvoaqmzmxswzanviiawjakoaatdbmvphcawsmravuif' \
  --header 'Content-Type: application/json' \
  --data '{
  "stream": true,
  "model": "qwen2.5-coder-14b",
  "messages": [
    {
      "role": "user",
      "content": "What opportunities and challenges will the Chinese large model industry face in 2025?"
    }
  ],
  "max_tokens": 1000
}'