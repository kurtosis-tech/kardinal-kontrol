from flask import Flask, render_template, request, redirect, url_for
import redis
import os

app = Flask(__name__)

redis_server = os.environ["REDIS"]

# Initialize Redis
r = redis.Redis(host=redis_server, port=6379)


if "OPTION1" in os.environ and os.environ["OPTION1"]:
    option1 = os.environ["OPTION1"]
else:
    option1 = "Option 1"

if "OPTION2" in os.environ and os.environ["OPTION2"]:
    option2 = os.environ["OPTION2"]
else:
    option2 = "Option 2"

if "TITLE" in os.environ and os.environ["TITLE"]:
    title = os.environ["TITLE"]
else:
    title = "Title"

# Set up initial vote counts
# TODO: implement this on redis proxy
# if not r.exists("option1"):
#     r.set("option1", 0)
# if not r.exists("option2"):
#     r.set("option2", 0)


@app.route("/", methods=["GET", "POST"])
def index():
    if request.method == "POST":
        vote = request.form["vote"]
        if vote == "option1":
            r.incr("option1")
        elif vote == "option2":
            r.incr("option2")
        return redirect(url_for("index"))

    # Get current vote counts
    option1_votes = int(r.get("option1"))
    option2_votes = int(r.get("option2"))
    return render_template(
        "index.html",
        option1_votes=option1_votes,
        option2_votes=option2_votes,
        subtitle=title,
        option1=option1,
        option2=option2,
    )


if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=80)
