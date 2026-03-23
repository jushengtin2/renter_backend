terraform {
  cloud {
    organization = "hazukashii" 
    workspaces {
      name = "renter_backend"
    }
  }
}