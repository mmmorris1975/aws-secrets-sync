require 'spec_helper'

describe 'tests using command line args' do

  describe 'for the ssm backend using default key' do
    describe command ('aws-secrets-sync -s ssm H4sIAG/UnlwAA6tW0k/OLErOSU3O1E+vyixQslJQKkktLlGq5QIATeaDghsAAAA=') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/gzip$/ }
    end

    describe command ('aws-secrets-sync -s ssm eyIvY2lyY2xlY2kvYmFzZTY0IjogInZhbHVlIn0K') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/base64/ }
    end

    describe command ('aws-secrets-sync -s ssm \'{"/circleci/text": "item"}\'') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/text/ }
    end

    describe command ('aws-secrets-sync -s ssm -o /circleci/one-shot boom') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/one-shot/ }
    end

    describe command ('aws-secrets-sync -s ssm H4sIAH7PnlwAA6tW0k9KTFGyUlDKKs7PU+ICADjxdc8QAAAA') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /ERROR error decoding json: / }
    end

    describe command ('aws-secrets-sync -s ssm eyIvbm9hY2Nlc3MvYjY0IjogInRlc3QifQo=') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /ERROR error storing secret: AccessDeniedException: / }
    end
  end

  describe 'for the ssm backend using custom key' do
    describe command ('aws-secrets-sync -s ssm -k alias/circleci H4sIAG/UnlwAA6tW0k/OLErOSU3O1E+vyixQslJQKkktLlGq5QIATeaDghsAAAA=') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/gzip$/ }
    end

    describe command ('aws-secrets-sync -s ssm -k alias/circleci eyIvY2lyY2xlY2kvYmFzZTY0IjogInZhbHVlIn0K') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/base64/ }
    end

    describe command ('aws-secrets-sync -s ssm -k alias/circleci \'{"/circleci/text": "item"}\'') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/text/ }
    end

    describe command ('aws-secrets-sync -s ssm -k alias/circleci -o /circleci/one-shot boom') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/one-shot/ }
    end
  end

  describe 'for the dynamodb backend' do
    describe command ('aws-secrets-sync -s dynamodb H4sIAG/UnlwAA6tW0k/OLErOSU3O1E+vyixQslJQKkktLlGq5QIATeaDghsAAAA=') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /FATAL missing required table name for dynamodb backend$/ }
    end

    describe command ('aws-secrets-sync -s dynamodb -t test-table H4sIAG/UnlwAA6tW0k/OLErOSU3O1E+vyixQslJQKkktLlGq5QIATeaDghsAAAA=') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /FATAL failed to lookup KMS key/ }
    end

    describe command ('aws-secrets-sync -s dynamodb -t test-table -k alias/circleci H4sIAG/UnlwAA6tW0k/OLErOSU3O1E+vyixQslJQKkktLlGq5QIATeaDghsAAAA=') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/gzip$/ }
    end

    describe command ('aws-secrets-sync -s dynamodb -t test-table -k alias/circleci eyIvY2lyY2xlY2kvYmFzZTY0IjogInZhbHVlIn0K') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/base64/ }
    end

    describe command ('aws-secrets-sync -s dynamodb -t test-table -k alias/circleci \'{"/circleci/text": "item"}\'') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/text/ }
    end

    describe command ('aws-secrets-sync -s dynamodb -t test-table -k alias/circleci -o /circleci/one-shot boom') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/one-shot/ }
    end

    describe command ('aws-secrets-sync -s dynamodb -t test-table -k alias/circleci H4sIAH7PnlwAA6tW0k9KTFGyUlDKKs7PU+ICADjxdc8QAAAA') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /ERROR error decoding json: / }
    end

    describe command ('aws-secrets-sync -s dynamodb -t terraform-locks -k alias/circleci H4sIAG/UnlwAA6tW0k/OLErOSU3O1E+vyixQslJQKkktLlGq5QIATeaDghsAAAA=') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /FATAL error describing dynamodb table: AccessDeniedException: / }
    end
  end

  describe 'for the s3 backend' do
    describe command ('aws-secrets-sync -s s3 H4sIAG/UnlwAA6tW0k/OLErOSU3O1E+vyixQslJQKkktLlGq5QIATeaDghsAAAA=') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /FATAL missing required bucket name for s3 backend$/ }
    end

    describe command ('aws-secrets-sync -s s3 -b trash-686784119290 H4sIAG/UnlwAA6tW0k/OLErOSU3O1E+vyixQslJQKkktLlGq5QIATeaDghsAAAA=') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /FATAL failed to lookup KMS key/ }
    end

    describe command ('aws-secrets-sync -s s3 -b trash-686784119290 -k alias/circleci H4sIAG/UnlwAA6tW0k/OLErOSU3O1E+vyixQslJQKkktLlGq5QIATeaDghsAAAA=') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/gzip$/ }
    end

    describe command ('aws-secrets-sync -s s3 -b trash-686784119290 -k alias/circleci eyIvY2lyY2xlY2kvYmFzZTY0IjogInZhbHVlIn0K') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/base64/ }
    end

    describe command ('aws-secrets-sync -s s3 -b trash-686784119290 -k alias/circleci \'{"/circleci/text": "item"}\'') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/text/ }
    end

    describe command ('aws-secrets-sync -s s3 -b trash-686784119290 -k alias/circleci -o /circleci/one-shot boom') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/one-shot/ }
    end

    describe command ('aws-secrets-sync -s s3 -b trash-686784119290 -k alias/circleci H4sIAH7PnlwAA6tW0k9KTFGyUlDKKs7PU+ICADjxdc8QAAAA') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /ERROR error decoding json: / }
    end

    describe command ('aws-secrets-sync -s s3 -b terraform-state-686784119290 -k alias/circleci H4sIAG/UnlwAA6tW0k/OLErOSU3O1E+vyixQslJQKkktLlGq5QIATeaDghsAAAA=') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /ERROR error storing secret: AccessDenied: / }
    end
  end

  describe 'for the secrets manager backend' do
    if ENV.fetch("CIRCLECI", false).to_s === "false"; then
      describe command ('aws-secrets-sync -s secretsmanager H4sIALF6n1wAA6tWSs4sSs5JTc7UT6/KLFCyUlAqSS0uUarlAgBdWbyoGgAAAA==') do
        its(:exit_status) { should eq 0 }
        its(:stderr) { should match /INFO updated secret circleci\/gzip$/ }
      end

      describe command ('aws-secrets-sync -s secretsmanager eyJjaXJjbGVjaS9iYXNlNjQiOiAidmFsdWUifQo=') do
        its(:exit_status) { should eq 0 }
        its(:stderr) { should match /INFO updated secret circleci\/base64/ }
      end

      describe command ('aws-secrets-sync -s secretsmanager \'{"circleci/text": "item"}\'') do
        its(:exit_status) { should eq 0 }
        its(:stderr) { should match /INFO updated secret circleci\/text/ }
      end

      describe command ('aws-secrets-sync -s secretsmanager -o circleci/one-shot boom') do
        its(:exit_status) { should eq 0 }
        its(:stderr) { should match /INFO updated secret circleci\/one-shot/ }
      end
    end

    describe command ('aws-secrets-sync -s secretsmanager H4sIAH7PnlwAA6tW0k9KTFGyUlDKKs7PU+ICADjxdc8QAAAA') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /ERROR error decoding json: / }
    end

    describe command ('aws-secrets-sync -s ssm eyIvbm9hY2Nlc3MvYjY0IjogInRlc3QifQo=') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /ERROR error storing secret: AccessDeniedException: / }
    end
  end
end