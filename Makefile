
all: windows linux mac 

windows:
	go env -w CGO_ENABLED=0
	go env -w GOOS=windows
	go build -o gis.exe cmd/main.go

linux:
	go env -w CGO_ENABLED=0
	go env -w GOOS=linux
	go build -o gis_linux cmd/main.go

mac:
	go env -w CGO_ENABLED=0
	go env -w GOOS=darwin
	go build -o gis_mac cmd/main.go