require 'spec_helper'

describe 'tests using command line args' do

  describe 'for the ssm backend using default key' do

    describe command ('aws-secrets-sync -s ssm H4sIABrNnlwAA6tW0k+vyixQslJQKkktLlGq5QIANZyavxIAAAA=') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/gzip$/ }
    end

    describe command ('aws-secrets-sync -s ssm eyIvYmFzZTY0IjogInZhbHVlIn0K') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/base64/ }
    end

    describe command ('aws-secrets-sync -s ssm \'{"/text": "item"}\'') do
      its(:exit_status) { should eq 0 }
      its(:stderr) { should match /INFO updated secret \/text/ }
    end

    describe command ('aws-secrets-sync -s ssm H4sIAH7PnlwAA6tW0k9KTFGyUlDKKs7PU+ICADjxdc8QAAAA') do
      its(:exit_status) { should eq 1 }
      its(:stderr) { should match /ERROR error decoding json: / }
    end
  end
end