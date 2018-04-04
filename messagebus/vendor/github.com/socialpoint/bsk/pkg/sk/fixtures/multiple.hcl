service geoip {
	description = "IP address to geo location converter"

	operation locate {
		description = "Given an IP address find the location"

		input ip {
			type     = "string"
			required = true
		}

		output {
			location {
				coutry {
					type = "string"
				}
				city {
					type = "string"
				}
				latitude {
					type = "string"
				}
			}
		}
	}
}


service crosspromotion {
	description = "Cross promotion service"

	operation get {
		description = "Get cross promotion settings"

		input {
			fake {
				type    = "boolean"
				default = false
			}


			platform {
				type     = "string"
				required = true
			}

			user_id {
				type     = "string"
				required = true
			}

			apps_installed {
				type     = "string"
				required = true
			}
		}

		output {
			xpromo {
				id {
					type = "string"
				}

				banners {
					type      = "list"
					structure = "banner"
				}

				check_apps {
					type      = "list"
					structure = "string"
				}
			}
		}

		errors {
			invalid_application {
				description = "One of the applications provided is not valid"
			}

			invalid_platform {
				description = "The platform provided is not valid"
			}

			invalid_game {
				description = "The game provided is not valid"
			}
		}
	}
}

structure xpromo {

}

structure banner {
	game {

	}

	app_id {
	}

	button {
	}

	store_id {

	}

	background {

	}

	current {
		type = "boolean"
	}
}
