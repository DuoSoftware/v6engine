package main
import(
  "fmt"
  "encoding/json"
)

func main() {
  input :="{\"AppType\":\"HTML5SDK\",\"AppUri\":\"\",\"ApplicationID\":\"APP_SHELL_DEVSTUDIO\",\"Description\":\"Developer Tools\",\"ImageId\":\"\",\"Name\":\"Developer Tools\",\"SecretKey\":\"2538336539450d17791c5b47e89030e9\",\"iconUrl\":\"/devportal/appicons/APP_SHELL_DEVSTUDIO.png\",\"timestamp\":null}";;
  m := make(map[string]interface{})
  err := json.Unmarshal([]byte(input), &m)
  if err != nil {
      fmt.Println(err)
  }
  fmt.Println(m["AppType"])
  fmt.Println(m["ApplicationID"])
}
