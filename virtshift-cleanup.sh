#!/bin/bash

set -x

if [ -z "$1" ]; then
	echo  "usage: %0 <POOL> (virsh pool-list)"
	exit 1
fi
POOL="${1}"

for DOMAIN in $(virsh -c "${CONNECT}" list --all --name)
do
	virsh -c "${CONNECT}" destroy "${DOMAIN}"
	virsh -c "${CONNECT}" undefine "${DOMAIN}"
done

virsh -c "${CONNECT}" vol-list "${POOL}" | tail -n +3 | while read -r VOLUME _
do
	if test -z "${VOLUME}"
	then
		continue
	fi
	virsh -c "${CONNECT}" vol-delete --pool "${POOL}" "${VOLUME}"
done
virsh -c "${CONNECT}" pool-delete ${POOL}
virsh -c "${CONNECT}" pool-destroy ${POOL}
virsh -c "${CONNECT}" pool-undefine ${POOL}

for NET in $(virsh -c "${CONNECT}" net-list --all --name)
do
	if test "${NET}" = default
	then
		continue
	fi
	virsh -c "${CONNECT}" net-destroy "${NET}"
	virsh -c "${CONNECT}" net-undefine "${NET}"
done

