Go resources
====

 * [Go Dependencies](#godeps)
 * [Books](#books)
    * [Starter Books](#starter-books)
 * [Other Resources](#resources)

**Go Dependencies**
=====
Added vendor and godep dirs.

Steps - 
```
go get github.com/tools/godep
go get github.com/Azure/Guest-Configuration-Extension
cd $GOPATH/github.com/Azure/Guest-Configuration-Extension
# Some dependencies for godep save - 
go get github.com/kr/logfmt
go get github.com/davecgh/go-spew/spew
go get github.com/pmezard/go-difflib/difflib
godep save ./main/
```

`godep save ./main/` generates the vendor and the godeps dir. 

Usage - https://github.com/tools/godep#edit-test-cycle

```
godep go build
godep go install
godep go test
```

**Books**
=====

**Starter Books**
----

### [The Little Go Book](http://openmymind.net/The-Little-Go-Book/) *Free*

### [An Introduction to Programming in Go](http://www.golang-book.com/) *Free*

### [Go Bootcamp](http://www.golangbootcamp.com/) *Free*

### [Learning Go](http://www.miek.nl/go) *Free*

### [Introducing Go](https://www.safaribooksonline.com/library/view/introducing-go/9781491941997/) *Available on Safari with Microsoft SSO*

**Other Resources**
=====

### [Godep: Dependency Management in Go](https://blog.codeship.com/godep-dependency-management-in-golang/)

### [Godep: Dependency Management in Go](https://blog.codeship.com/godep-dependency-management-in-golang/)
