service basic {
	description = "Basic service definition"

	operation "greet" {
		description = "Greet by name"
		method      = "GET"

		input {
			length {
				type        = "integer"
				required    = true
				description = "the length"

				cli {
					position = 1
				}
			}

			name {
				type        = "string"
				required    = true
				description = "the name"

				cli {
					position = 0
				}
			}
		}

		output "greetings" {
			type        = "string"
			description = "the greetings"
		}
	}
}