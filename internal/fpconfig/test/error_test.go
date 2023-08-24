package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser"
)

func TestStepErrorConfig(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/http.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.bad_http"] == nil {
		assert.Fail("bad_http pipeline not found")
		return
	}

}

func TestStepErrorConfigRetries(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/error.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.bad_http_retries"] == nil {
		assert.Fail("bad_http_retries pipeline not found")
		return
	}

	step := pipelines["local.pipeline.bad_http_retries"].GetStep("http.my_step_1")

	if step == nil {
		assert.Fail("step not found")
		return
	}

	errorConfig := step.GetErrorConfig()
	if step == nil {
		assert.Fail("error config not found")
		return
	}

	assert.Equal(2, errorConfig.Retries)
}
