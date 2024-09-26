

transform {
  use "js" {    
    script_path = "/tmp/js_script.js"
  }
}

target {
  use "http" {
    # URL endpoint
    url = "https://acme.com/x"
  }
}