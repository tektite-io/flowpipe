pipeline "for_loop" {

    param "users" {
        type = list(string)
        default = ["jerry", "Janis", "Jimi"]
    }

    step "echo" "text_1" {
        for_each = param.users
        text = "user if ${each.value}"
    }

    step "echo" "no_for_each" {
        text = "baz"
    }
}

pipeline "for_depend_object" {

    param "users" {
        type = list(string)
        default = ["brian", "freddie", "john", "roger"]
    }

    step "echo" "text_1" {
        for_each = param.users
        text = "user if ${each.value}"
    }

    step "echo" "text_3" {
        for_each = step.echo.text_1
        text = "output one value is ${each.value.text}"
    }
}


pipeline "for_loop_depend" {

    param "users" {
        type = list(string)
        default = ["jerry", "Janis", "Jimi"]
    }

    step "echo" "text_1" {
        for_each = param.users
        text = "user is ${each.value}"
    }

    step "echo" "text_2" {
        text = "output is ${step.echo.text_1[0].text}"
    }

    step "echo" "text_3" {
        for_each = step.echo.text_1
        text = "output one value is ${each.value.text}"
    }
}