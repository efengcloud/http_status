all:
	sudo rpmbuild -D '_sourcedir '`pwd` -D '_rpmdir '`pwd` -D "_release $*" -bb *.spec

