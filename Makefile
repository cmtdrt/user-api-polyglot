# Indicate to Make that the following are targets, not files
.PHONY: rust go 

rust:
	docker compose up rust

go:
	docker compose up go
