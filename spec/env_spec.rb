require 'spec_helper'

describe 'tests using environment variables' do

  describe 'for the ssm backend using default key' do

    before(:each) do
      ENV['SECRETS_BACKEND']='ssm'
      ENV.delete('KMS_KEY')
    end

    after(:each) do
      ENV.delete('SECRETS_BACKEND')
    end

    describe command ('aws-secrets-sync H4sIAG/UnlwAA6tW0k/OLErOSU3O1E+vyixQslJQKkktLlGq5QIATeaDghsAAAA=') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/gzip$/ }
    end

    describe command ('aws-secrets-sync eyIvY2lyY2xlY2kvYmFzZTY0IjogInZhbHVlIn0K') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/base64/ }
    end

    describe command ('aws-secrets-sync \'{"/circleci/text": "item"}\'') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/text/ }
    end

    describe command ('aws-secrets-sync H4sIAH7PnlwAA6tW0k9KTFGyUlDKKs7PU+ICADjxdc8QAAAA') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /ERROR error decoding json: / }
    end

    describe command ('aws-secrets-sync eyIvbm9hY2Nlc3MvYjY0IjogInRlc3QifQo=') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /ERROR error storing secret: AccessDeniedException: / }
    end

    describe 'using one-shot mode' do

      before(:each) do
        ENV['ONE_SHOT']='true'
      end

      after(:each) do
        ENV.delete('ONE_SHOT')
      end

      describe command ('aws-secrets-sync /circleci/one-shot boom') do
        its(:exit_status) { should eq 0 }
        its(:stderr) { should match /INFO updated secret \/circleci\/one-shot/ }
      end
    end
  end

  describe 'for the ssm backend using custom key' do

    before(:each) do
      ENV['SECRETS_BACKEND']='ssm'
      ENV['KMS_KEY']='alias/circleci'
    end

    after(:each) do
      ENV.delete('SECRETS_BACKEND')
      ENV.delete('KMS_KEY')
    end

    describe command ('aws-secrets-sync H4sIAG/UnlwAA6tW0k/OLErOSU3O1E+vyixQslJQKkktLlGq5QIATeaDghsAAAA=') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/gzip$/ }
    end

    describe command ('aws-secrets-sync eyIvY2lyY2xlY2kvYmFzZTY0IjogInZhbHVlIn0K') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/base64/ }
    end

    describe command ('aws-secrets-sync \'{"/circleci/text": "item"}\'') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/text/ }
    end

    describe 'using one-shot mode' do

      before(:each) do
        ENV['ONE_SHOT']='1'
      end

      after(:each) do
        ENV.delete('ONE_SHOT')
      end

      describe command ('aws-secrets-sync /circleci/one-shot boom') do
        its(:exit_status) { should eq 0 }
        its(:stderr) { should match /INFO updated secret \/circleci\/one-shot/ }
      end
    end
  end

  describe 'for the ssm backend using advanced parameters' do
    before(:each) do
      ENV['SECRETS_BACKEND']='ssm'
      ENV['SSM_ADVANCED']='1'
      ENV['KMS_KEY']='alias/circleci'
    end

    after(:each) do
      ENV.delete('SECRETS_BACKEND')
      ENV.delete('SSM_ADVANCED')
      ENV.delete('KMS_KEY')
    end

    if ENV.fetch("CIRCLECI", false).to_s === "false"; then
      describe command ('aws-secrets-sync \'{"/circleci/advanced-text-env": "item"}\'') do
        its(:exit_status) { should eq 0 }
        its(:stderr) { should match /INFO updated secret \/circleci\/advanced-text-env/ }
      end
    end
  end

  describe 'for the dynamodb backend' do

    before(:each) do
      ENV['SECRETS_BACKEND']='dynamodb'
      ENV['DYNAMODB_TABLE']='test-table'
      ENV['KMS_KEY']='alias/circleci'
    end

    after(:each) do
      ENV.delete('SECRETS_BACKEND')
      ENV.delete('DYNAMODB_TABLE')
      ENV.delete('KMS_KEY')
    end

    describe command ('aws-secrets-sync H4sIAG/UnlwAA6tW0k/OLErOSU3O1E+vyixQslJQKkktLlGq5QIATeaDghsAAAA=') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/gzip$/ }
    end

    describe command ('aws-secrets-sync eyIvY2lyY2xlY2kvYmFzZTY0IjogInZhbHVlIn0K') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/base64/ }
    end

    describe command ('aws-secrets-sync \'{"/circleci/text": "item"}\'') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/text/ }
    end

    describe command ('aws-secrets-sync H4sIAH7PnlwAA6tW0k9KTFGyUlDKKs7PU+ICADjxdc8QAAAA') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /ERROR error decoding json: / }
    end

    describe 'using one-shot mode' do

      before(:each) do
        ENV['ONE_SHOT']='1'
      end

      after(:each) do
        ENV.delete('ONE_SHOT')
      end

      describe command ('aws-secrets-sync /circleci/one-shot boom') do
        its(:exit_status) { should eq 0 }
        its(:stderr) { should match /INFO updated secret \/circleci\/one-shot/ }
      end
    end
  end

  describe 'for the s3 backend' do

    before(:each) do
      ENV['SECRETS_BACKEND']='s3'
      ENV['S3_BUCKET']='trash-686784119290'
      ENV['KMS_KEY']='alias/circleci'
    end

    after(:each) do
      ENV.delete('SECRETS_BACKEND')
      ENV.delete('S3_BUCKET')
      ENV.delete('KMS_KEY')
    end

    describe command ('aws-secrets-sync H4sIAG/UnlwAA6tW0k/OLErOSU3O1E+vyixQslJQKkktLlGq5QIATeaDghsAAAA=') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/gzip$/ }
    end

    describe command ('aws-secrets-sync eyIvY2lyY2xlY2kvYmFzZTY0IjogInZhbHVlIn0K') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/base64/ }
    end

    describe command ('aws-secrets-sync \'{"/circleci/text": "item"}\'') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/circleci\/text/ }
    end

    describe command ('aws-secrets-sync H4sIAH7PnlwAA6tW0k9KTFGyUlDKKs7PU+ICADjxdc8QAAAA') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /ERROR error decoding json: / }
    end

    describe 'using one-shot mode' do

      before(:each) do
        ENV['ONE_SHOT']='1'
      end

      after(:each) do
        ENV.delete('ONE_SHOT')
      end

      describe command ('aws-secrets-sync /circleci/one-shot boom') do
        its(:exit_status) { should eq 0 }
        its(:stderr) { should match /INFO updated secret \/circleci\/one-shot/ }
      end
    end
  end

  describe 'for the secrets manager backend' do

    before(:each) do
      ENV['SECRETS_BACKEND']='secretsmanager'
    end

    after(:each) do
      ENV.delete('SECRETS_BACKEND')
    end

    if ENV.fetch("CIRCLECI", false).to_s === "false"; then
      describe command ('aws-secrets-sync H4sIALF6n1wAA6tWSs4sSs5JTc7UT6/KLFCyUlAqSS0uUarlAgBdWbyoGgAAAA==') do
        its(:exit_status) { should eq 0 }
        its(:stderr) { should match /INFO updated secret circleci\/gzip$/ }
      end

      describe command ('aws-secrets-sync eyJjaXJjbGVjaS9iYXNlNjQiOiAidmFsdWUifQo=') do
        its(:exit_status) { should eq 0 }
        its(:stderr) { should match /INFO updated secret circleci\/base64/ }
      end

      describe command ('aws-secrets-sync \'{"circleci/text": "item"}\'') do
        its(:exit_status) { should eq 0 }
        its(:stderr) { should match /INFO updated secret circleci\/text/ }
      end

      describe 'using one-shot mode' do

        before(:each) do
          ENV['ONE_SHOT']='1'
        end

        after(:each) do
          ENV.delete('ONE_SHOT')
        end

        describe command ('aws-secrets-sync circleci/one-shot boom') do
          its(:exit_status) { should eq 0 }
          its(:stderr) { should match /INFO updated secret circleci\/one-shot/ }
        end
      end
    end

    describe command ('aws-secrets-sync H4sIAH7PnlwAA6tW0k9KTFGyUlDKKs7PU+ICADjxdc8QAAAA') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /ERROR error decoding json: / }
    end

    describe command ('aws-secrets-sync eyIvbm9hY2Nlc3MvYjY0IjogInRlc3QifQo=') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /ERROR error storing secret: AccessDeniedException: / }
    end
  end
end