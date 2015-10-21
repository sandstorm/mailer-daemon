# Concept for Mailing System

* Input Data
  * A JSON file with receivers, in a way where each line is a JSON object describing one person. The way this JSON is
    structured (per line) is arbitrary.

* A "Receiver Group" is a subset of the JSON file of receivers.
  * the idea is to use the "JQ" tool on the command line to filter the receiver lists.

* See the "mailing-concept.graffle" file.



    Constraint: We want to be sure that emails are really sent exactly once.
        That's why it is not enough to just have a queue as "dispatching mechanism" and inside the worker, check if the mail is already sent. Because if a mail is popped from the queue, and the worker dies before sending, it would not be discovered.
        It is quite difficult to know whether a worker died or not. That's why we need leases for that.
        We need some kind of persistent data structure, that's why we will use redis.
        We'll slightly tune the system such that a few emails might be sent twice in case of error; because this cannot be prevented but it is very unlikely. We think this is better than a few people which do not receive emails at all.
    In the longer run, we want to be able to parallelize the sending of a newsletter across multiple systems.
    Depending on the selected Recipient list, there exist different placeholders which can be used in the contents of the email.
    TYPO3 Neos has a new document type "Newsletter"
        Property: Template (Select-List)
        Property: Recipient List
        Property: Email Subject
        Open Question: Scheduled Newsletter or Click-To-Send?
        Content of the newsletter is the content in Neos.
        The content contains placeholder like "$receiverName" which the user has to insert as text.
            Validation: Using JavaScript, we will check whether the used placeholders are all defined. If not, display an error.
            Preview: Optionally, we could add a "preview" which is triggered once the user switches to the Neos preview mode; then replacing the placeholders by example values.
    Redis Data Structure Concept – the following data structures are created PER NEWSLETTER which shall be sent.
        SET unassigned
            contains JSON with the receiver email address and all placeholder values.
            is filled by the "Receiver List Reader", which takes the data from CSV files or API and puts it in here
        LIST inProgress
            contains the same json as in "unassigned"
        STRING leases:[task] 1  [leaseTimeout]
            for each element inside inProgress, a lease is created by adding such a string with a timeout.
        LIST done
            contains the same json as in "unassigned", if a job is really finished and sent it is moved from inProgress to here.
    To send a newsletter, we have two parts in the system:
        Receiver List Reader: read the receiver list; and add entries into "unassigned"
        Sender (will spawn multiple sending-threads)
            WHILE unassigned and inProgress is not empty
                inside a single transaction/lua script:
                    remove a random element from the receiver list
                    element to inProgress list. remember index of where we inserted this.
                    create lease
                start extra goroutine which will take care of increasing the lease timeout while the sender sends the email
                try to send email; replacing the placeholders in HTML.
                    TODO: how should the sending occur? SMTP or proprietary API?
                if DONE, do the following inside the transaction:
                    remove lease
                    remove element from inProgress list, add element to "done" list.
                    // At this point, a slight chance of re-sending an email occurs, if the system crashed after sending, but without removing the elements from inProgress list.
            try to work through inProgress and see if any leases have expired. If so, take over these jobs.
    Bounce Handling
        This is still an open question on how to do this, because SMTP sending is asynchronous; thus bounces are delivered to a certain email address. If using an API such as Mandrill (see https://mandrillapp.com/api/docs/messages.JSON.html), we get bounce information directly; but if using SMTP, that is more difficult.
