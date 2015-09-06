## procfs

Package system provide some information that could be retrieved by scanning the `procfs` on linux. It's entirely native and requires no dependencies.

TODO:
* faire un environnement de test plus complet
* faire le has bound port and stuff
* faire un outil identique à ps et netstat (en plus simple certes)
* benchmarquer le truc (voir comment faire des benchs etc ...)
* voir si mes allocs de array sont optimales
* godoc
* hacker news
* etc ...

* integration test sur gops


gops -> show only the user process (user, pid, ppid, name)
gops -a -> show all processes (user, pid, ppid, name) including non user processes
gops -p -> show socket and port bound (user, pid, ppid, name, [proto:port, proto:port, ...])
gops pid

global option: a, p

pid 
