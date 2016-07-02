#!/bin/bash
#
# script to make sql file for inserting all "currently trusted" keys
#

root=$(readlink -e "$(dirname "$0")")
touch "$root/keys.sql"
for key in $(cat "$root/keys.txt") ; do
	echo "insert into modprivs(pubkey, newsgroup, permission) values('$key', 'overchan', 'all');" >> keys.sql ;
done
