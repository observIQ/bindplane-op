describe user('bindplane') do
    it { should exist }
    its('uid') { should eq 10001 }
    its('group') { should eq 'bindplane' }
    its('lastlogin') { should eq nil }
end

describe file('/bindplane') do
    its('mode') { should cmp '0755' }
    its('owner') { should eq 'root' }
    its('group') { should eq 'root' }
    its('type') { should cmp 'file' }
end

[
    "data/storage",
].each do |dir|
    describe file(dir) do
        its('mode') { should cmp '0640' }
        its('owner') { should eq 'bindplane' }
        its('group') { should eq 'bindplane' }
        its('type') { should cmp 'file' }
    end
end

describe port(3001) do
    it { should be_listening }
    its('processes') {should include 'bindplane'}
    # should never be udp or udp6, will be tcp or
    # tcp6 depending on the test system
    its('protocols') { should_not include('udp') }
    its('protocols') { should_not include('udp6') }
end

describe processes('bindplane') do
    its('users') { should eq ['bindplane'] }
end
