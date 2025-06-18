build:
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/ws ./cmd/server

host_ip_prd = '178.156.172.188'

deploy: build
	rsync -P ./bin/ws dodo@${host_ip_prd}:~
	ssh -t dodo@${host_ip_prd} '\
	  sudo systemctl enable ws \
	  && sudo systemctl restart ws \
	  '