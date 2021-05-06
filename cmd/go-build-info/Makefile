# do:
# 	rm -f ./$${PWD##*/}
# 	make -C . all
# 	go build
# 	echo;echo;echo
# 	- ./$${PWD##*/}
# 	go build
# 	echo;echo;echo
# 	- ./$${PWD##*/}

all: bootstrap
	./$${PWD##*/} ~/go/active/bin/go build -ldflags '-w -s'
	./$${PWD##*/}
bootstrap:
	if [[ ! -e $${PWD##*/} ]]; then ~/go/active/bin/go build -ldflags '-w -s'; fi

install: 
	if [[ ! -e $${PWD##*/} ]]; then		\
  if [[ -n $$(which $${PWD##*/}) ]]; then	\
    $${PWD##*/}  ~/go/active/bin/go install -ldflags '-w -s';	\
  else						\
     ~/go/active/bin/go install -ldflags '-w -s';		\
  fi;						\
else						\
  $${PWD##*/}  ~/go/active/bin/go install -ldflags '-w -s';	\
fi
