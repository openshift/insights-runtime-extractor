.PHONY: golang-app quarkus-3.13.0 quarkus-3.15.1 spring-boot RHBQ plain-java

all: quarkus-3.13.0 quarkus-3.15.1 spring-boot golang-app RHBQ plain-java

golang-app:
	cd ./golang-app && make

quarkus-3.13.0:
	cd ./quarkus/3.13.0. && make

quarkus-3.15.1:
	cd ./quarkus/3.15.1 && make

spring-boot:
	cd ./spring-boot && make

RHBQ:
	cd ./RHBQ && make

plain-java:
	cd ./plain-java && make
