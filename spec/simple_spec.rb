require 'spec_helper'

describe command ('aws-secrets-sync -h') do
  its(:exit_status) { should eq 2 }
  its(:stderr) { should match /^Usage of .{,2}aws-secrets-sync:/ }
end

describe command ('aws-secrets-sync -V') do
  its(:exit_status) { should eq 1 }
  its(:stderr) { should match /VERSION: \d+\.\d+\.\d+(-\d+-\w+)?/ }
end

describe command ('aws-secrets-sync') do
  its(:exit_status) { should eq 1 }
  its(:stderr) { should match /FATAL backend  is not valid, must be one of: / }
end