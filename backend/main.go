package main

import "backend/cmd"

func main() {
	cmd.Run(nil)

	// to run in embeded mode:
	// cmd.Run(
	// 	cmd.WithPort(8082),
	// 	cmd.WithDefaultUser("admin@localhost.com", "admin"),
	// 	cmd.WithDefaultProject("Backend", "go", "backend-dev-token"),
	// 	cmd.WithDefaultProject("Frontend", "sveltekit", "frontend-dev-token"),
	// )
}
