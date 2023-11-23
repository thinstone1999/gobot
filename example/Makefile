
env=linux
cmd=release
.PHONY:build
build:
	cd cmd && go build -o app.exe

.PHONY:rebuild
rebuild:stop log build start

.PHONY:restart
restart:stop log start

.PHONY:start
start:
	cd cmd/script && python start.py

.PHONY:stop
stop:
	cd cmd/script && python stop.py

.PHONY:pb
pb:
	cd cmd/script && python pb.py

.PHONY:log
log:
	cd cmd/script && python log.py

.PHONY:app
app:
	cd cmd/app && go run main.go

.PHONY:release
release:
	cd cmd/script && python release.py -env=${env} -cmd=${cmd}

.PHONY:linux
linux:
	cd cmd &&  GOOS=linux CGO_ENABLED=0 go build -o appx
	docker cp cmd/appx centos:/tree
	docker cp cmd/cache/app.json centos:/tree
	docker cp cmd/cache/robot.b3 centos:/tree