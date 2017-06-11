$gocmd='github.com/deviceio/agent/cmd/deviceio-agent'

$env:GOOS='windows';$env:GOARCH='amd64';go install std; go build -o $env:GOPATH\bin\deviceio-agent.windows.amd64.exe $gocmd
$env:GOOS='windows';$env:GOARCH='386';go install std; go build -o $env:GOPATH\bin\deviceio-agent.windows.386.exe $gocmd 

$env:GOOS='linux';$env:GOARCH='amd64';go install std; go build -o $env:GOPATH\bin\deviceio-agent.linux.amd64 $gocmd
$env:GOOS='linux';$env:GOARCH='386';go install std; go build -o $env:GOPATH\bin\deviceio-agent.linux.386 $gocmd
$env:GOOS='linux';$env:GOARCH='ppc64';go install std; go build -o $env:GOPATH\bin\deviceio-agent.linux.ppc64 $gocmd
$env:GOOS='linux';$env:GOARCH='ppc64le';go install std; go build -o $env:GOPATH\bin\deviceio-agent.linux.ppc64le $gocmd
$env:GOOS='linux';$env:GOARCH='mips';go install std; go build -o $env:GOPATH\bin\deviceio-agent.linux.mips $gocmd
$env:GOOS='linux';$env:GOARCH='mipsle';go install std; go build -o $env:GOPATH\bin\deviceio-agent.linux.mipsle $gocmd
$env:GOOS='linux';$env:GOARCH='mips64';go install std; go build -o $env:GOPATH\bin\deviceio-agent.linux.mips64 $gocmd
$env:GOOS='linux';$env:GOARCH='mips64le';go install std; go build -o $env:GOPATH\bin\deviceio-agent.linux.mips64le $gocmd