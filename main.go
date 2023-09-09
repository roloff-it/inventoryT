package main

func main() {
	app := App{}
	app.Initialize(DBUser, DBPassword, DBName)
	app.Run("192.168.99.36:10000")
}
