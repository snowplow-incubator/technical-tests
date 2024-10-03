

transform {
  use "js" {    
    script_path = "/tmp/js_script.js"
  }
}

target {
  use "http" {
    # URL endpoint. Change this as appropriate for your tests.
    url = "https://example.com/x"
  }
}
