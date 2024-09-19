.PHONY: golang-app upstream-quarkus spring-boot

all: upstream-quarkus spring-boot golang-app

golang-app: setup
	cd ./golang-app && make

upstream-quarkus: setup
	cd ./upstream-quarkus && make

spring-boot: setup
	cd ./spring-boot && make

setup:
	./setup.sh
