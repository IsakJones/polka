prep:
	./scripts/prep.sh $(n)

stnone:
	./scripts/start_without.sh

stlog:
	./scripts/start_with_logs.sh 

stmux:
	./scripts/start_with_tmux.sh 

clean:
	./scripts/clean_up.sh

reset-db:
	./scripts/reset_db.sh

docker_prep:
	./scripts/docker_prep.sh

spam:
	./scripts/check_gen_env.sh
	./generator/bin/polkagenerator -w=$(w) -t=$(t)

spamm:
	./scripts/check_gen_env.sh
	./generator/bin/polkagenerator -w=$(w) -t=$(t) -m

getsnap:
	./generator/bin/polkagenerator -gs

settle:
	./generator/bin/polkagenerator -sb

view-logs:
	# Display logs with multitail
	multitail -s 2 -sn 1,2 cache/log.txt \
						balancer/log.txt \
						receiver/node0/log.txt
