import os


os.system('redis-server &')
os.system('go run main.go &')
os.system('curl -sS  -H "Content-Type: application/json" --data @data.json http://localhost:8080/')
os.system('redis-cli shutdown')
