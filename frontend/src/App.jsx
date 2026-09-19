import {
  HashRouter,
  Routes,
  Route,
  Link,
  useParams,
} from "react-router-dom";

import { useEffect, useState } from "react";
import "./App.css";

function Home() {
  const isLoggedIn = !!localStorage.getItem("token");

  return (
    <div className="app">
      <nav className="navbar">
        <Link to="/" className="logo">
          Live<span>Poll</span>
        </Link>

        <div className="nav-links">
          {isLoggedIn ? (
            <button
              className="logout-button"
              onClick={() => {
                localStorage.removeItem("token");
                window.location.reload();
              }}
            >
              Logout
            </button>
          ) : (
            <>
              <Link to="/login">Login</Link>

              <Link to="/signup" className="nav-button">
                Sign Up
              </Link>
            </>
          )}
        </div>
      </nav>

      <main className="hero">
        <div className="live-badge">
          <span></span> LIVE POLLING
        </div>

        <h1>
          Ask. Vote.
          <br />
          <span>See it live.</span>
        </h1>

        <p>
          Create interactive polls, share them with anyone,
          and watch the results update instantly.
        </p>

        <div className="hero-buttons">
          {isLoggedIn ? (
            <Link to="/create" className="primary-button">
              Create a Poll →
            </Link>
          ) : (
            <>
              <Link to="/login" className="primary-button">
                Login to Create a Poll
              </Link>

              <Link to="/signup" className="secondary-button">
                Create Account
              </Link>
            </>
          )}
        </div>

        <div className="features">
          <div>
            <strong> Real-time</strong>
            <span>Instant vote updates</span>
          </div>

          <div>
            <strong> Shareable</strong>
            <span>Share with anyone</span>
          </div>

          <div>
            <strong> Secure</strong>
            <span>Protected poll creation</span>
          </div>
        </div>
      </main>
    </div>
  );
}

function Login() {
  const handleLogin = async (event) => {
    event.preventDefault();

    const email = event.target.email.value;
    const password = event.target.password.value;

    try {
      const response = await fetch("https://livepoll-8vit.onrender.com/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          email,
          password,
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        alert(data.error || "Login failed");
        return;
      }

      localStorage.setItem("token", data.token);

      window.location.href = "/";
    } catch {
      alert("Could not connect to backend");
    }
  };

  return (
    <div className="auth-page">
      <Link to="/" className="auth-logo">
        Live<span>Poll</span>
      </Link>

      <div className="auth-card">
        <h1>Welcome back</h1>

        <p>Login to create and manage your polls.</p>

        <form onSubmit={handleLogin}>
          <label>Email</label>

          <input
            name="email"
            type="email"
            placeholder="you@example.com"
            required
          />

          <label>Password</label>

          <input
            name="password"
            type="password"
            placeholder="Enter your password"
            required
          />

          <button
            type="submit"
            className="primary-button full"
          >
            Login
          </button>
        </form>

        <p className="auth-footer">
          Don't have an account?{" "}
          <Link to="/signup">Create one</Link>
        </p>
      </div>
    </div>
  );
}

function Signup() {
  const handleSignup = async (event) => {
    event.preventDefault();

    const name = event.target.name.value;
    const email = event.target.email.value;
    const password = event.target.password.value;

    try {
      const response = await fetch("https://livepoll-8vit.onrender.com/signup", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name,
          email,
          password,
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        alert(data.error || "Signup failed");
        return;
      }

      alert("Account created successfully!");

      window.location.href = "/login";
    } catch {
      alert("Could not connect to backend");
    }
  };

  return (
    <div className="auth-page">
      <Link to="/" className="auth-logo">
        Live<span>Poll</span>
      </Link>

      <div className="auth-card">
        <h1>Create account</h1>

        <p>Start creating live polls in seconds.</p>

        <form onSubmit={handleSignup}>
          <label>Name</label>

          <input
            name="name"
            type="text"
            placeholder="Your name"
            required
          />

          <label>Email</label>

          <input
            name="email"
            type="email"
            placeholder="you@example.com"
            required
          />

          <label>Password</label>

          <input
            name="password"
            type="password"
            placeholder="Minimum 6 characters"
            required
            minLength="6"
          />

          <button
            type="submit"
            className="primary-button full"
          >
            Create Account
          </button>
        </form>

        <p className="auth-footer">
          Already have an account?{" "}
          <Link to="/login">Login</Link>
        </p>
      </div>
    </div>
  );
}

function CreatePoll() {
  const [creating, setCreating] = useState(false);

  const handleCreatePoll = async (event) => {
    event.preventDefault();

    const question = event.target.question.value;

    const options = [
      event.target.option1.value,
      event.target.option2.value,
    ];

    const token = localStorage.getItem("token");

    if (!token) {
      alert("Please login first.");
      window.location.href = "/login";
      return;
    }

    setCreating(true);

    try {
      const response = await fetch("https://livepoll-8vit.onrender.com/polls", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          question,
          options,
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        alert(data.error || "Could not create poll");
        return;
      }

      window.location.href = `/#/poll/${data.poll.id}`;
    } catch {
      alert("Could not connect to backend");
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className="auth-page">
      <Link to="/" className="auth-logo">
        Live<span>Poll</span>
      </Link>

      <div className="poll-card create-card">
        <div className="live-badge">
          <span></span> CREATE POLL
        </div>

        <h1>Create a new poll</h1>

        <p className="card-description">
          Ask a question and let your audience vote in real time.
        </p>

        <form onSubmit={handleCreatePoll}>
          <label>Your question</label>

          <input
            name="question"
            type="text"
            placeholder="What do you want to ask?"
            required
          />

          <label>Answer options</label>

          <input
            name="option1"
            type="text"
            placeholder="Option 1"
            required
          />

          <input
            name="option2"
            type="text"
            placeholder="Option 2"
            required
          />

          <button
            type="submit"
            className="primary-button full"
            disabled={creating}
          >
            {creating ? "Creating..." : "Create Poll →"}
          </button>
        </form>
      </div>
    </div>
  );
}

function PollPage() {
  const { id } = useParams();

  const [poll, setPoll] = useState(null);
  const [counts, setCounts] = useState({});
  const [loading, setLoading] = useState(true);
  const [voting, setVoting] = useState(false);
  const [voted, setVoted] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const loadPoll = async () => {
      try {
        const pollResponse = await fetch(
          `https://livepoll-8vit.onrender.com/polls/${id}`
        );

        const pollData = await pollResponse.json();

        if (!pollResponse.ok) {
          alert(pollData.error || "Could not load poll");
          return;
        }

        setPoll(pollData.poll);

        const resultsResponse = await fetch(
          `https://livepoll-8vit.onrender.com/polls/${id}/results`
        );

        const resultsData = await resultsResponse.json();

        if (resultsResponse.ok) {
          setCounts(resultsData.counts);
        }
      } catch {
        alert("Could not connect to backend");
      } finally {
        setLoading(false);
      }
    };

    loadPoll();
  }, [id]);

  useEffect(() => {
    if (!id) return;

    const socket = new WebSocket(
      `wss://livepoll-8vit.onrender.com/polls/${id}/live`
    );

    socket.onmessage = (event) => {
      const update = JSON.parse(event.data);

      setCounts((currentCounts) => ({
        ...currentCounts,
        [update.option]: update.count,
      }));
    };

    return () => {
      socket.close();
    };
  }, [id]);

  const handleVote = async (option) => {
    if (voted) {
      return;
    }

    setVoting(true);

    try {
      const response = await fetch(
        `https://livepoll-8vit.onrender.com/polls/${id}/vote`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            option,
          }),
        }
      );

      const data = await response.json();

      if (!response.ok) {
        alert(data.error || "Vote failed");
        return;
      }

      setVoted(true);
    } catch {
      alert("Could not connect to backend");
    } finally {
      setVoting(false);
    }
  };

  const copyLink = async () => {
    await navigator.clipboard.writeText(window.location.href);

    setCopied(true);

    setTimeout(() => {
      setCopied(false);
    }, 2000);
  };

  if (loading) {
    return (
      <div className="center-page">
        <div className="loader"></div>

        <p>Loading poll...</p>
      </div>
    );
  }

  if (!poll) {
    return (
      <div className="center-page">
        <h2>Poll not found</h2>

        <Link to="/">Back to Home</Link>
      </div>
    );
  }

  const totalVotes = Object.values(counts).reduce(
    (sum, count) => sum + count,
    0
  );

  return (
    <div className="poll-page">
      <nav className="navbar">
        <Link to="/" className="logo">
          Live<span>Poll</span>
        </Link>

        <div className="live-status">
          <span></span> Live results
        </div>
      </nav>

      <main className="poll-container">
        <div className="poll-header">
          <div className="live-badge">
            <span></span> LIVE POLL
          </div>

          <h1>{poll.question}</h1>

          <p>
            {totalVotes}{" "}
            {totalVotes === 1 ? "vote" : "votes"} • Results update instantly
          </p>
        </div>

        <div className="options">
          {poll.options.map((option) => {
            const count = counts[option] || 0;

            const percentage =
              totalVotes === 0
                ? 0
                : Math.round((count / totalVotes) * 100);

            return (
              <button
                key={option}
                className={`option-card ${
                  voted ? "voted-option" : ""
                }`}
                onClick={() => handleVote(option)}
                disabled={voting || voted}
              >
                <div className="option-top">
                  <span>{option}</span>

                  <strong>{percentage}%</strong>
                </div>

                <div className="progress-track">
                  <div
                    className="progress-fill"
                    style={{
                      width: `${percentage}%`,
                    }}
                  ></div>
                </div>

                <div className="option-bottom">
                  <span>{count} votes</span>

                  {!voted && <span>Click to vote</span>}
                </div>
              </button>
            );
          })}
        </div>

        {voted && (
          <div className="success-message">
            ✓ Your vote has been submitted!
          </div>
        )}

        <div className="share-box">
          <div>
            <strong>Share this poll</strong>

            <p>Send this link to your audience.</p>
          </div>

          <button
            onClick={copyLink}
            className="copy-button"
          >
            {copied ? "✓ Copied!" : "Copy Link"}
          </button>
        </div>

        <Link to="/" className="back-link">
          ← Back to LivePoll
        </Link>
      </main>
    </div>
  );
}

function App() {
  return (
    <HashRouter>
      <Routes>
        <Route path="/" element={<Home />} />

        <Route path="/login" element={<Login />} />

        <Route path="/signup" element={<Signup />} />

        <Route path="/create" element={<CreatePoll />} />

        <Route path="/poll/:id" element={<PollPage />} />
      </Routes>
    </HashRouter>
  );
}

export default App;