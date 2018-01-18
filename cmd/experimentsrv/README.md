# Introduction

The experiment server is used to persist experiment details and to record changes to the state of experiments.  Items included within an experiment include layer definitions and meta-data items.

The experiment server offers a gRPC API that can be accessed using a machine-to-machine or human-to-machine (HCI) interface.  The HCI interface can be interacted with using the grpc_cli tool provided with the gRPC toolkita  More information about grpc_cli can be found at, https://github.com/grpc/grpc/blob/master/doc/command_line_tool.md.

# Experiment Database

The experiment server makes use of a Postgres DB.  The installation process is specific to AWS Aurora.  To begin installation you will need to create a Postgres Aurora RDS instance.  Use values of your choosing for the DB name and user/password combinations.

Parameters that impact deployment of your Aurora instance include, RDS Endpoint, DB Name, user name, and the password.

## Installation

Now go to the experimentsrv.yaml file and change the Egress rule to point at your endpoint.  The deployment spec, PGHOST and PGDATABASE should also be modified to the endpoint and the DB Name respectively.

You should now create or edit the secrets.yaml file ready for deployment with the user name and the password.

# Service Authentication

Service authetication is explained within the top level README.md file for the github.com/SentientTechnologies/platform-services repository.  All calls into the experiment service must contain metadata for the autorization brearer token and have the all:experiments claim in order to be accepted.

# Manually exercising the server

```
root@experiments-v1-bc46b5d68-dqnwk:~# /tmp/grpc_cli call localhost:30001 ai.sentient.experiment.ExperimentService.Get "id: ''" --metadata authorization:"$AUT
H"                                                                                                                                                            
connecting to localhost:30001
Sending client initial metadata:
authorization : eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6IlJqYzRSRUV5TmpFM09UQXdSRFZCTmtSQ056QkNPVEJETURrelFVUkZNRFk0UmpsRFJVTXhNUSJ9.eyJpc3MiOiJodHRwczovL3NlbnRpZW50YWkuYXV0aDAuY29tLyIsInN1YiI6ImF1dGgwfDVhNWY4MjM2M2VjYTYxMGJkNjVjMzUwMyIsImF1ZCI6Imh0dHA6Ly9hcGkuc2VudGllbnQuYWkvZXhwZXJpbWVudHNydiIsImlhdCI6MTUxNjMxMzY0OCwiZXhwIjoxNTE2NDAwMDQ4LCJhenAiOiI3MWVMTnU5QncxcmdmWXo5UEEyZ1o0Smk3dWptM1V3aiIsInNjb3BlIjoiYWxsOmV4cGVyaW1lbnRzIiwiZ3R5IjoicGFzc3dvcmQifQ.UOAD8Epu93Excnhkt6x062LEXI0UO1b5ZV68j8h3ok1OrLHl6zI1iBgB___Xr_wWeTPBPSIL3zYEZXYcSFeJkQT0SkwBILi2iN6fNeOTohryY9vRjLQlfYyCOR2O2tbP9mPs6Mnn-PoAI9Tq93U0WGRYs90a46ICGfy23xhs7jZqpqocRa2A_EcAVzNJNX1xx8MIyFk-upMQutvcpJ_H7YxfMpDHaPY9jMRDs7tSz9lD0-10PnKAwpo3PyA54s4vP76uaxDVET5T9RLyxKUScL_g2CP_fDk2dF2_DnLG9rWPe4mP_yPble09ejKUdB3DEVHm1Ia8dkrVUFw00B73Dg
experiment {
  createTime {
    seconds: 1516319288
  }
}

Rpc succeeded with OK status
root@experiments-v1-bc46b5d68-dqnwk:~# export AUTH="eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6IlJqYzRSRUV5TmpFM09UQXdSRFZCTmtSQ056QkNPVEJETURrelFVUkZNRFk0UmpsRFJVTXhNUSJ9.eyJpc3MiOiJodHRwczovL3NlbnRpZW50YWkuYXV0aDAuY29tLyIsInN1YiI6ImF1dGgwfDVhNWY4MjM2M2VjYTYxMGJkNjVjMzUwMyIsImF1ZCI6Imh0dHA6Ly9hcGkuc2VudGllbnQuYWkvZXhwZXJpbWVudHNydiIsImlhdCI6MTUxNjMxMzY0OCwiZXhwIjoxNTE2NDAwMDQ4LCJhenAiOiI3MWVMTnU5QncxcmdmWXo5UEEyZ1o0Smk3dWptM1V3aiIsInNjb3BlIjoiYWxsOmV4cGVyaW1lbnRzIiwiZ3R5IjoicGFzc3dvcmQifQ.UOAD8Epu93Excnhkt6x062LEXI0UO1b5ZV68j8h3ok1OrLHl6zI1iBgB___Xr_wWeTPBPSIL3zYEZXYcSFeJkQT0SkwBILi2iN6fNeOTohryY9vRjLQlfYyCOR2O2tbP9mPs6Mnn-PoAI9Tq93U0WGRYs90a46ICGfy23xhs7jZqpqocRa2A_EcAVzNJNX1xx8MIyFk-upMQutvcpJ_H7YxfMpDHaPY9jMRDs7tSz9lD0-10PnKAwpo3PyA54s4vP76uaxDVET5T9RLyxKUScL_g2CP_fDk2dF2_DnLG9rWPe4mP_yPble09ejKUdB3DEVHm1Ia8dkrVUFw00B73Dg"
root@experiments-v1-bc46b5d68-dqnwk:~# /tmp/grpc_cli call localhost:30001 ai.sentient.experiment.ExperimentService.Get "id: ''" --metadata authorization:"$AUTH"
connecting to localhost:30001
Sending client initial metadata:
authorization : eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6IlJqYzRSRUV5TmpFM09UQXdSRFZCTmtSQ056QkNPVEJETURrelFVUkZNRFk0UmpsRFJVTXhNUSJ9.eyJpc3MiOiJodHRwczovL3NlbnRpZW50YWkuYXV0aDAuY29tLyIsInN1YiI6ImF1dGgwfDVhNWY4MjM2M2VjYTYxMGJkNjVjMzUwMyIsImF1ZCI6Imh0dHA6Ly9hcGkuc2VudGllbnQuYWkvZXhwZXJpbWVudHNydiIsImlhdCI6MTUxNjMxMzY0OCwiZXhwIjoxNTE2NDAwMDQ4LCJhenAiOiI3MWVMTnU5QncxcmdmWXo5UEEyZ1o0Smk3dWptM1V3aiIsInNjb3BlIjoiYWxsOmV4cGVyaW1lbnRzIiwiZ3R5IjoicGFzc3dvcmQifQ.UOAD8Epu93Excnhkt6x062LEXI0UO1b5ZV68j8h3ok1OrLHl6zI1iBgB___Xr_wWeTPBPSIL3zYEZXYcSFeJkQT0SkwBILi2iN6fNeOTohryY9vRjLQlfYyCOR2O2tbP9mPs6Mnn-PoAI9Tq93U0WGRYs90a46ICGfy23xhs7jZqpqocRa2A_EcAVzNJNX1xx8MIyFk-upMQutvcpJ_H7YxfMpDHaPY9jMRDs7tSz9lD0-10PnKAwpo3PyA54s4vP76uaxDVET5T9RLyxKUScL_g2CP_fDk2dF2_DnLG9rWPe4mP_yPble09ejKUdB3DEVHm1Ia8dkrVUFw00B73Dg
experiment {
  createTime {
    seconds: 1516319308
  }
}

Rpc succeeded with OK status

```

