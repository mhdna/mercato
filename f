BODY="{"code": "I-$f", "name": "T-Shirt", "description": "A nice t-shirt", "price": 50000, "discount": 0, "color_id": 1, "size_id": 1, "attributes": { "0": "clothing", "1": "nike", "2": "men" } } "

for f in { 1 .. 200 }; do
 curl -X POST http://localhost:8077/products \
  -H "Content-Type: application/json" \
  -d ''

done

for f in { 1 .. 200 }; do
 curl -X POST http://localhost:8077/products \
  -H "Content-Type: application/json" \
  -d "{
    \"code\": \"I-$f\",
    \"name\": \"T-Shirt\",
    \"description\": \"A nice t-shirt\",
    \"price\": 50000,
    \"discount\": 0,
    \"color_id\": 1,
    \"size_id\": 1,
    \"attributes\": {
      \"0\": \"clothing\",
      \"1\": \"nike\",
      \"2\": \"men\"
    }
  }"
done
