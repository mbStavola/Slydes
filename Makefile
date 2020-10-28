# Force separator to spaces instead of tabs
.RECIPEPREFIX := $(.RECIPEPREFIX) # Is this comment useless or...? :)

build:
    go build -o slydes main.go debug.go

test:
    go test ./...