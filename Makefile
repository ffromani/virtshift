all: virtshift-install

virtshift-install:
	$(MAKE) -C installer all
	mv installer/cmd/virtshift-install/virtshift-install .

clean:
	rm -f virtshift-install
