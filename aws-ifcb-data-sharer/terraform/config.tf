terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.72.2"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "3.6.1"
    }
  }


  required_version = ">= 1.5.7"
}

provider "aws" {
  region = "us-east-1"
}
