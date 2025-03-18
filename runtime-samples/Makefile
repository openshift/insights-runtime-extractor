.PHONY: golang-app quarkus-3.13.0 quarkus-3.15.1 spring-boot-3.1.4 spring-boot-3.4.3 RHBQ plain-java

all: quarkus-3.13.0 quarkus-3.15.1 spring-boot-3.1.4 spring-boot-3.4.3 golang-app RHBQ plain-java

golang-app:
	cd ./golang-app && make

quarkus-3.13.0:
	cd ./quarkus/3.13.0. && make

quarkus-3.15.1:
	cd ./quarkus/3.15.1 && make

spring-boot-3.1.4:
	cd ./spring-boot/3.1.4 && make

spring-boot-3.4.3:
	cd ./spring-boot/3.4.3 && make

RHBQ:
	cd ./RHBQ && make

plain-java:
	cd ./plain-java && make
