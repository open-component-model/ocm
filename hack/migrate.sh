#!/bin/bash -e

PACKAGE="github.com/open-component-model/ocm"

build=(components .github Dockerfile flake.nix Makefile .goreleaser.yaml dist/config.yaml README.md .reuse/dep5)

adapts=( api pkg cmds hack examples )


addMigrator()
{
  if [ -n "$MIGRATOR" ]; then
     echo "$NESTED" "$@" >>"${MIGRATOR}"
  fi
  if [ -n "$SCRIPT" ]; then
    echo "$@"
  fi
}

package()
{
  addMigrator package "$1"

  PACKAGE="$1"
  prefix="$PACKAGE/"
}

remove()
{
  for i in "$@"; do
    if [ -f "$i" -o -d "$i" ]; then
       echo "removing $i"
       rm -rf "$i"
    else
       echo "already removed $i"
    fi
  done
}

subst()
{
   echo -n "s:$1\([^a-zA-Z0-9]\):$2\1:g"
}

substDot()
{
   echo -n "s:\./$1\([^a-zA-Z0-9]\):\./$2\1:g"
}


adaptMig()
{
  local dst="$2/$(basename "$1")"

  addMigrator adaptMig "$1" "$2"

  for s in "${adapts[@]}" "${@:3}"; do
    if [ -e "$s" ]; then
      echo "${NESTED}  adapting absolute $s"
      find "$s" -type f -exec sed -i "$(subst "$prefix$1" "$prefix$dst")" {} \;
    fi
  done
}

mig()
{
  local P
  if [ "$1" == "-p" ]; then
    P="$1"
    shift
  fi
  local dst="$2/$(basename "$1")"
  if [ -d "$dst" ]; then
    echo "${NESTED}already migrated $1 to $dst"
    return
  fi
  if [ ! -d "$1" ]; then
    echo "${NESTED}already migrated $1 to $dst"
    return
  fi
  echo "${NESTED}migrate $1 to $dst"

  mkdir -p "$2"
  mv "$1" "$2"

  adaptMig "$1" "$2" "${@:3}"

  # adapting relative (file) paths
  if [ -z "$P" ]; then
    for s in cmds examples "${@:3}"; do
      if [ -e "$s" ]; then
        echo "${NESTED}  adapting relative $s"
        find "$s" -type f -exec sed -i "$(subst "$1" "$dst")" {} \;
      fi
    done
 else
    for s in cmds examples "${@:3}"; do
      if [ -e "$s" ]; then
        echo "${NESTED}  adapting relative $s"
        find "$s" -type f -exec sed -i "$(substDot "$1" "$dst")" {} \;
      fi
    done
  fi
}

substDirect()
{
  # -r requires unescaped brackets
  echo -n "s:(import\s+)\"$1\":\1\"$2\":g"
}

substLabeled()
{
  # -r requires unescaped brackets
  echo -n "s:([a-zA-Z0-9_]+\s+)\"$1\":\1\"$2\":g"
}

substUnLabeled()
{
  local old="$(basename "$1")"
  # -r requires unescaped brackets
  echo -n "s:\"$1\":$old \"$2\":g"
}

move()
{
  if [ -d "$2" ]; then
    echo "already moved $1 to $2"
    return
  fi
  if [ ! -d "$1" ]; then
    echo "already moved $1 to $2"
    return
  fi
  
  if [ "$(basename "$1")" == "$(basename "$2")" ]; then
    mig "$1" "$(dirname "$2")"
    return
  fi

  if [ "$(dirname "$1")" == "$(dirname "$2")" ]; then
    rename "$1" "$(basename "$2")"
    return
  fi

  mig "$1" "$(dirname "$2")"
  rename "$(dirname "$2")/$(basename "$1")" "$(basename "$2")"
}

adaptRename()
{
  local dst="$(dirname "$1")/$2"

  addMigrator adaptRename "$@"

  for s in "${adapts[@]}"; do
    if [ -e "$s" ]; then
      echo "  adapting $s"
      find "$s" -type f -exec sed -r -i -e "$(substDirect "$prefix$1" "$prefix$dst")" -e "$(substLabeled "$prefix$1" "$prefix$dst")" -e "$(substUnLabeled "$prefix$1" "$prefix$dst")" {} \;
    fi
  done
}

adaptRenameParent()
{
  local dst="$(dirname "$1")/$2"

  addMigrator adaptRenameParent "$@"

  for s in "${adapts[@]}" "${@:3}"; do
    if [ -e "$s" ]; then
      echo "  adapting $s"
      find "$s" -type f -exec sed -i "$(subst "$1" "$dst")" {} \;
    fi
  done
}

rename() {
  local dst="$(dirname "$1")/$2"

  if [ -d "$dst" ]; then
    echo already renamed "$1" to "$dst"
    return
  fi
  if [ ! -d "$1" ]; then
    echo already renamed "$1" to "$dst"
    return
  fi

  echo "rename $1 to $dst"

  mkdir -p "$dst"
  for d in "$1"/*; do
    if [ -d "$d" ]; then
      if [ "$d" != testdata ]; then
         NESTED="  " mig "$d" "$dst"
      fi
    fi
  done

  local files=( "$1"/* )
  if [ ${#files[@]} -ne 1 -o "${files[0]}" != "$1/*" ]; then
    mv "$1"/* "$dst"

    adaptRename "$@"
  fi

  adaptRenameParent "$@"

  rmdir "$1"
}

fixRelativePathsFor()
{
  p="$1"
  dst=
  lvl=1

  while true; do
    p="$p/*"
    (( lvl++ ))

    files=( $p )

    if [ ${#files[@]} -eq 1 -a "${files[0]}" == "$p" ]; then
      break
    fi

    echo "level $lvl: $p"
    dst="$dst../"
    
    d="$(sed 's/\./\\\./g' <<<"$dst")"

    for f in "${files[@]}"; do
      if [ -f "$f" -a ! -h "$f" ]; then
        for t in "${@:2}"; do
          sed -r -i -e "s:([^./])(\.\./)+$t:\1$d$t:g" "$f"
        done
      fi
    done
  done
}


changeModuleName()
{
  if [ -n "$MODULE" ]; then
    echo "*** changing module name"

    addMigrator MODULE=\""$MODULE"\"
    addMigrator changeModuleName

    for s in "${adapts[@]}" "$@" go.mod .golangci.yaml; do
      find "$s" -type f -exec sed -i -e "s:$PACKAGE:$MODULE:g" -e "s:\(LABEL.*\)$MODULE:\1$PACKAGE:g" -e "s&\([uU]rl:.*\)$MODULE&\1$PACKAGE&g" {} \;
    done

    #echo "*** adapting nix description"
    #sed -i -e "s=github:open-component-model/ocm=github:open-component-model/ocm-core=g" README.md
  fi
}

##########################################################
# main
##########################################################

if [ "$1" == "--reset" ]; then
  echo "*** reset"
  git checkout HEAD .
  rm -rf api
  rm -rf cmds/ocm/common
  exit
fi

mkdir -p api

if [ "$1" == "--test" ]; then
  mig pkg/contexts/ocm api "${build[@]}"
  sed -i -e "s/pkg/api/g" Makefile
  sed -i -e "s/pkg/api/g" components/*/Dockerfile
  exit
fi

while [ $# -gt 0 ]; do
  case "$1" in
    --migrator)
      MIGRATOR="$2"
      shift 2;;
    --script) 
      SCRIPT="$2"
      shift 2;;
    --paths) 
      FIXPATHS=X
      shift;;
    --module)
      MODULE=ocm.software/ocm
      shift;;
    --deprecated)
      DEPRECATED=X
      shift;;
    -*)
      echo "invalid option $1" >&2
      exit 1;;
    *)
      if [ -z "$SCRIPT" ]; then
        echo "invalid extra arguments $@" >&2
        exit 1
      fi
      break;;
  esac
done

package "$PACKAGE"

if [ -n "$SCRIPT" ]; then
  MIGRATOR=
  adapts=( "${@}" )
  if [ ${#adapts[@]} -eq 0 ]; then
    adapts=( . )
  fi

  if [ ! -f "$SCRIPT" ]; then
    echo "invalid script $SCRIPT" >&2
    exit 1
  fi
  echo migrating "${adapts[@]}" with script "$SCRIPT"
  source "$SCRIPT"
  exit 0
fi

if [ -n "$MIGRATOR" ]; then
  echo "restructure packages and create migration script..."
else
  echo "restructure packages..."
fi

echo "*** removing deprecated packages..."
remove pkg/{contexts/options,errors,finalizer,exception,generics,optionutils,regex,tokens} pkg/utils/{pkgutils,testutils} testdata
remove pkg/helm/identity pkg/contexts/oci/identity

if [ -n "$DEPRECATED" ]; then
  exit 0
fi

echo "*** relocating packages..."
# contexts
for s in datacontext clictx config credentials oci; do
  mig pkg/contexts/$s api
done
rename api/clictx cli

mig pkg/contexts/ocm api "${build[@]}"
mig api/ocm/registration api/ocm/plugin

# technology support
for s in helm npm maven signing; do
  mig pkg/$s api/tech
done
mig pkg/docker api/tech .golangci.yaml

rename cmds/ocm/pkg common

# utils
mig pkg/utils api
mig -p pkg/runtime api/utils
for s in pkg/{spiff,dirtree,blobaccess,clisupport,cobrautils,encrypt,errkind,filelock,mime,out,refmgmt,iotools,listformat,logging,registrations,runtimefinalizer,semverutils,testutils} pkg/common/{accessio,accessobj,compression}; do
  mig $s api/utils
done

move pkg/common api/utils/misc

for s in pkg/{env/builder,env}; do
  mig $s api/helper
done


# now migrate sub packaged from migrated packages

# config extensions
for s in api/config/config; do
  mig $s api/config/extensions
done

# credentials extensions
for s in api/credentials/repositories; do
  mig $s api/credentials/extensions
done

# oci extensions
for s in api/oci/{repositories,actions,attrs}; do
  mig $s api/oci/extensions
done

# oci tools support
for s in api/oci/transfer; do
  mig $s api/oci/tools
done

# ocm extensions
for s in api/ocm/{repositories,accessmethods,blobhandler,attrs,download,actionhandler,digester,labels,resourcetypes,pubsub}; do
  mig $s api/ocm/extensions "${build[@]}"
done

# ocm tools support
for s in pkg/toi api/ocm/{signing,transfer}; do
  mig $s api/ocm/tools
done

mig pkg/version api "${build[@]}"

echo "*** renaming packages..."
rename api/ocm/context types
rename api/ocm/utils ocmutils
rename api/ocm/extensions/resourcetypes artifacttypes

echo "*** adapting go source root folders"
sed -i -e "s/pkg/api/g" Makefile
sed -i -e "s/pkg/api/g" components/*/Dockerfile

if [ -n "$FIXPATHS" ]; then
  echo "*** adapting relative paths"
  for i in api cmds examples docs; do
    fixRelativePathsFor "$i" api cmds examples docs resources
  done
fi

if [ -n "$MODULE" ]; then
  changeModuleName components docs "${build[@]}" go.mod .golangci.yaml
fi

echo "*** formatting"
go fmt ./{api,cmds}/... >/dev/null

echo "*** generating"
make generate
