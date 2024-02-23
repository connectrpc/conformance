# Testing Servers

Lorem ipsum dolor sit amet convectateur


Create a separate markdown page for "Testing a server" , to describe the process of implementing a server-under-test.

The page does not need to regurgitate the details that are in the protos. Instead, it should outline the protocol:
- reading from stdin, writing results to stdout, writing other log messages to stderr), provide some pseudo-code for how to handle each of the methods of the conformance service, and point to the relevant doc comments via linking to the generated docs in the BSR. 

The server will likely be the shorter doc (shorter than a client guide), since it has less interaction with the test runner and only handles a single configuration per process.

It should link to the "Configuring Conformance Tests" page, about how to configure the features that the server supports and how to define known failing and known flaky test cases.

It could also link to examples in the connect-es repo.
