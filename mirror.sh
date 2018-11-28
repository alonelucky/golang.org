#!/bin/bash

url="https://github.com/golang/"
urls=(
    "arch"
	"build"
	"crypto"
	"debug"
	"dl"
	"exp"
	"image"
	"lint"
	"mobile"
	"net"
	"perf"
	"playground"
	"review"
	"scratch"
	"sync"
	"sys"
	"time"
	"tools"
	"tour"
	"vgo"
)
str="/archive/master.zip"

for u in ${urls[@]};do
  echo $url$u$str;
  wget $url$u$str
  unzip master.zip
  mv $u-master $u
  rm master.zip
done;