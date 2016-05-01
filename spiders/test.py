from twisted.internet import defer, protocol, reactor

test = reactor.spawnProcess(protocol.ProcessProtocol(), "scrapy", ("scrapy", "runspider", "investigate.py"))

print test.pid
print test

reactor.run()
