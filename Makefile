run-browser:
	cd browser-client && npm run dev

run-signaling:
	cd signaling-server && go run main.go

run-server:
	cd runner-client && go run main.go
