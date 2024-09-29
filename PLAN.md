Just want to write some general ramblings on what I'm thinking to immortilise my ideas and prevent them being washed away by new waves of ideas :)

# Serving the BT application 

- Do we NEED to use bubbletea or do we WANT to use bubbletea. At the moment wish ssh server and 
bubbletea seems best used with the MiddlewareWithProgramHandler which dictates we server the app
immediately upon connection - I guess this does kind of make sense.

My current idea is to have the ssh server basically silo/group connections.

So as connections are made, we group them into fours (assuming four player games only and maybe we add bots as standard until all someone starts the game I dont know yet really im always in pvp mode)

We could then have a game server inside the common folder which is started by the ssh server with incremental or preset addresses - say it was a tcp server just increment the port by 1 for every four connections 

What we really want then is for the bt application to be created with the relevant server in mind. 

This way, for every 5th application a new server is spawned and the following four players will connect to the new server?!

So its like:

Client connect -> ssh -> 1 -> start tcp 2 -> start and serve bubbletea game #

I think this way like all ssh is now worried about is serving a useful application to the user with predefined connection details etc - I dunno maybe this will work 


# hanlders etc

we have to satisfy a few interfaces to make bt and wish ssh server work. #

The most important are handlers really which is just a func(ssh.session) (tea.Model, []tea.ProgramOption)

This handler can be used to hook into the ssh middleware and execute things on connection.

This can differ from a ProgramHandler that takes a session and returns a program. 

So I guess we can have like an order of handlers, and one should be once conn it made, set-up ready for the bt programme to start, then inside the programme handler we input the server details and off we go I think?!=




# sat 28.09

We have ssh middleware, and inside our server is a list of all connections. We can communicate with all the current connections using a p.Send().

So I think what we want to try is to make a connection, and add the name of connections to each persons screen as a new person connects to the server now. s