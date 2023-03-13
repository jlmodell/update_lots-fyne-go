# update_lots-fyne-go
- go build -o lot_updater.exe ./main.go  
- ./lot_updater.exe  

## uses internal api
- app relies on internally hosted api (fastapi POST to /update)  
- all processing handled on server (sends csv to server and displays the response)  
- additionally, logs to log file
