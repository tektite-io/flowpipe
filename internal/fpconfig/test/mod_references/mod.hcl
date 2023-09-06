

mod "pipeline_with_references" {
    title = "Test mod"
    description = "Use this mod for testing references within pipeline and from one pipeline to another"
}


pipeline "foo" {

    # leave this here to ensure that references that is later than the resource can be resolved
    #
    # we parse the HCL files from top to bottom, so putting this step `baz` after `bar` is the easier path
    # reversing is the a harder parse
    step "echo" "baz" {
        text = step.echo.bar
    }

    step "echo" "bar" {
        text = "test"
    }

    step "pipeline" "child_pipeline" {
        pipeline = pipeline.foo_two
    }

    step "echo" "child_pipeline" {
        text = step.pipeline.child_pipeline.foo
    }
}


pipeline "foo_two" {
    step "echo" "baz" {
        text = "foo"
    }

    outout "foo" {
        value = echo.baz.text
    }
}