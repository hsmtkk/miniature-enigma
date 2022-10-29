// Copyright (c) HashiCorp, Inc
// SPDX-License-Identifier: MPL-2.0
import { Construct } from "constructs";
import { App, TerraformStack, CloudBackend, NamedCloudWorkspace } from "cdktf";
import * as google from '@cdktf/provider-google';

const location = 'asia-northeast1';
const project = 'miniature-enigma';
const openweather_key_secret_id = 'openweather-key';

class MyStack extends TerraformStack {
  constructor(scope: Construct, id: string) {
    super(scope, id);

    new google.GoogleProvider(this, 'Google', {
      project,
    });

    new google.ArtifactRegistryRepository(this, 'docker_registry', {
      format: 'DOCKER',
      location,
      repositoryId: 'docker',      
    });

    const run_service_account = new google.ServiceAccount(this, 'run_service_account', {
      accountId: 'run-service-account',
    });

    const public_policy = new google.DataGoogleIamPolicy(this, 'public_policy', {
      binding: [{
        members: ['allUsers'],
        role: 'roles/run.invoker',
      }],
    });

    const back_run = new google.CloudRunService(this, 'back_run', {
      autogenerateRevisionName: true,
      location,
      metadata: {
        annotations: {
          'run.googleapis.com/ingress': 'internal',
        },
      },
      name: 'back',
      template: {
        spec: {
          containers: [{
            env: [{
              name: 'OPENWEATHER_KEY_SECRET_ID',
              value: openweather_key_secret_id,
            }],
            image: 'asia-northeast1-docker.pkg.dev/miniature-enigma/docker/back',
          }],
          serviceAccountName: run_service_account.email,
        },
      },
    });

    const front_run = new google.CloudRunService(this, 'front_run', {
      autogenerateRevisionName: true,
      location,
      name: 'front',
      template: {
        spec: {
          containers: [{
            env: [{
              name: 'COLLECTION',
              value: 'openweather',
            },{
              name: 'BACK_URL',
              value: back_run.status.get(0).url,
            }],
            image: 'asia-northeast1-docker.pkg.dev/miniature-enigma/docker/front',
          }],
          serviceAccountName: run_service_account.email,
        },
      },
    });

    new google.CloudRunServiceIamPolicy(this, 'back_policy', {
      location,
      service: back_run.name,
      policyData: public_policy.policyData,
    });

    new google.CloudRunServiceIamPolicy(this, 'front_policy', {
      location,
      service: front_run.name,
      policyData: public_policy.policyData,
    });

    new google.CloudbuildTrigger(this, 'build_trigger', {
      filename: 'cloudbuild.yaml',
      github: {
        owner: 'hsmtkk',
        name: project,
        push: {
          branch: 'main',
        },
      },
    });

    new google.SecretManagerSecret(this, 'secret_manager', {
      secretId: openweather_key_secret_id,
      replication: {
        automatic: true, 
      },
    });

    new google.CloudSchedulerJob(this, 'front_schedule', {
      name: 'front_schedule',
      httpTarget: {
        uri: front_run.status.get(0).url,
      },
      schedule: '* * * * *',
    });
  }
}

const app = new App();
const stack = new MyStack(app, "miniature-enigma");
new CloudBackend(stack, {
  hostname: "app.terraform.io",
  organization: "hsmtkkdefault",
  workspaces: new NamedCloudWorkspace("miniature-enigma")
});
app.synth();
