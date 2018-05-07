
部署节点
-----------------

第一步:
```
wget https://github.com/peatio/Sia/releases/download/peatio-v1.0.0rc1/siac-linux_amd64
```

```
wget https://github.com/peatio/Sia/releases/download/peatio-v1.0.0rc1/siad-linux_amd64
```

第二步: 
```
chmod +x siac-linux_amd64
```

```
chmod +x siad-linux_amd64
```

第三步:
```
sudo mv siac-linux_amd64 /usr/bin/siac
```

```
sudo mv siad-linux_amd64 /usr/bin/siad
```

第四步:
```
siad -M gct &
```

检查是否在同步:
```
siac
```

从源码编译

--------------------
<!---
To build from source, [Go 1.10 must be installed](https://golang.org/doc/install)
on the system. Make sure your `$GOPATH` is set, and then simply use `go get`:

```
go get -u github.com/peatio/Sia/...
```

This will download the Sia repo to your `$GOPATH/src` folder and install the
`siad` and `siac` binaries in your `$GOPATH/bin` folder.

To stay up-to-date, run the previous `go get` command again. Alternatively, you
can use the Makefile provided in this repo. Run `git pull origin master` to
pull the latest changes, and `make release` to build the new binaries. You
can also run `make test` and `make test-long` to run the short and full test
suites, respectively. Finally, `make cover` will generate code coverage reports
for each package; they are stored in the `cover` folder and can be viewed in
your browser.
-->