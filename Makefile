all: SHELL:=/bin/bash
all: release?=1
all:
	sudo rpmbuild -D '_sourcedir '`pwd` -D '_rpmdir '`pwd` -D '_builddir '`pwd` -D '_release '${release} -bb *.spec

