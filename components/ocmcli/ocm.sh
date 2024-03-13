#!/bin/sh
set -e

# this if will check if the first argument is a flag but only works if all arguments require a hyphenated flag -v; -SL; -f arg; etc will work, but not arg1 arg2
if [ "$#" -eq 0 ] || [ "${1#-}" != "$1" ]; then
    set -- "$@"
    exec /bin/ocm "$@"
    return
fi

# this case will check if the first argument is a known OCM command
case $1 in
  add | bootstrap | cache | check | clean | completion | controller | create | credentials | describe | download | execute | get | hash | help | install | oci | ocm | show | sign | toi | transfer | verify | version)
    exec /bin/ocm "$@"
    return
    ;;
esac

# else default to run whatever the user wanted like "bash" or "sh"
exec "$@"
