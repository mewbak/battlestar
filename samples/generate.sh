#!/bin/sh

# Check for needed utilities
which yasm >/dev/null || (echo 'Could not find yasm'; exit 1)
which ld >/dev/null || (echo 'Could not find ld'; exit 1)
which gcc >/dev/null || echo 'Could not find gcc (optional)'

battlestarc=../battlestarc
if [ ! -e $battlestarc ]; then
  make -C ..
fi

bits=`getconf LONG_BIT`
osx=$([[ `uname -s` = Darwin ]] && echo true || echo false)
asmcmd="yasm -f elf$bits"
ldcmd='ld -s --fatal-warnings -nostdlib --relax'
cccmd="gcc -Os -m64 -nostdlib"

if [[ $bits = 32 ]]; then
  ldcmd='ld -s -melf_i386 --fatal-warnings -nostdlib --relax'
  cccmd='gcc -Os -m32 -nostdlib'
fi

if [[ $osx = true ]]; then
  asmcmd='yasm -f macho'
  ldcmd='ld -macosx_version_min 10.8 -lSystem'
  bits=32
fi

for f in *.bts; do
  n=`echo ${f/.bts} | sed 's/ //'`
  echo "Building $n"
  # Don't output the log if "fail" is in the filename
  if [[ $n != *fail* ]]; then
    $battlestarc -bits="$bits" -osx="$osx" -f "$f" -o "$n.asm" -oc "$n.c" 2> "$n.log" || (cat "$n.log"; rm -f "$n.asm"; echo "$n failed to build!")
  else
    $battlestarc -bits="$bits" -osx="$osx" -f "$f" -o "$n.asm" -oc "$n.c" 2> "$n.log" || (rm -f "$n.asm"; echo "$n failed to build (correct)")
  fi
  [ -e $n.c ] && ($cccmd -c "$n.c" -o "${n}_c.o" || echo "$n failed to compile")
  [ -e $n.asm ] && ($asmcmd -o "$n.o" "$n.asm" || echo "$n failed to assemble")
  if [ -e ${n}_c.o -a -e $n.o ]; then
    $ldcmd "${n}_c.o" "$n.o" -o "$n" || echo "$n failed to link"
  elif [ -e $n.o ]; then
    $ldcmd "$n.o" -o "$n" || echo "$n failed to link"
  fi
  echo
done
