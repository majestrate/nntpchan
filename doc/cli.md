#Command-line interface

**srndv2** comes with a selection of command-line arguments for managing your node.

##Rebuild all thumbnails

To rebuild all thumbnails run:

    ./srndv2 tool rethumb
    
##Generate a new tripcode keypair (prints to stdout)

    ./srndv2 tool keygen

##Add a public key to moderation trust

Where `publickey` is the public key to be added.

    ./srndv2 tool mod add publickey
    
##Remove a public key from moderation trust

Where `publickey` is the public key to be removed.

    ./srndv2 tool mod del publickey

##Add a new NNTP user

Where `username` is the username and `password` is the user's password for the new uer.

    ./srndv2 tool nntp add-login username password

##Remove an existing NNTP user

Where `username` is the username of the user to be deleted.

    ./srndv2 tool nntp del-login username
