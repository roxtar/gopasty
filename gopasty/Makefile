SRC=gopasty.go
all: deploy

deploy: 
	appcfg.py update ..

test: 
	6g $(SRC) && dev_appserver.py ..
