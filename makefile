build:
	cd backend/cmd && go run *.go

productbuild: #run with "&" at the end
	nohup make > logs/$(fileName).out 2>&1
