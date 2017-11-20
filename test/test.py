import os


#os.system('redis-server &')
#os.system('go run main.go &')
os.system('curl -sS  -H "Content-Type: application/json" --data @data1.json http://localhost:8080/')
os.system('curl -sS  -H "Content-Type: application/json" --data @data2.json http://localhost:8080/')
os.system('curl -sS  -H "Content-Type: application/json" --data @data3.json http://localhost:8080/')
os.system('curl -sS  -H "Content-Type: application/json" http://localhost:8080/stats')
#os.system('redis-cli shutdown')
