//const signalingServerURL = "https://autodiscovery-signaling.app-builder-on-prem.net";
const signalingServerURL = "http://localhost:8080";

const config = {
  iceServers: [
    {
      // STUN servers are used to determine the public IP of the host. This is kinda like going to whatismyip.org.
      // The public IP is typically sent as an ICE candidate.
      urls: ["stun:stun1.l.google.com:19302", "stun:stun2.l.google.com:19302"],
    },
  ],
  iceCandidatePoolSize: 10,
};

// this gets set once a data channel is opened so other components can use the channel
let localChannel;

let pollRemoteCandidatesHandle;
let sessionInput = document.getElementById("session");
let sendBtn = document.getElementById("send-btn");
let clientBtn = document.getElementById("client-btn");

sendBtn.onclick = (event) => {
  log(`sending a message over the data channel`);
  localChannel.send("hi");
};

const pc = new RTCPeerConnection(config);

// WebRTC requires a transceiver for ICE candidates to be sent properly. Ideally we wouldn't need this...
pc.addTransceiver("audio");

pc.addEventListener("connectionstatechange", (event) => {
  log(`new connection state: ${pc.connectionState}`);
});

pc.addEventListener("datachannel", (event) => {
  localChannel = event.channel;
  event.channel.addEventListener("message", (event) => {
    const decodedMsg = new TextDecoder("utf-8").decode(event.data);
    log(`got message from PAR over data channel: ${decodedMsg}`);
  });

  localChannel.onopen = () => {
    sendBtn.disabled = false;
  };
});

clientBtn.onclick = async () => {
  clientBtn.disabled = true;

  log(`getting session`);
  let session = await getSession();

  log(`setting remote offer`);
  await pc.setRemoteDescription(session.offer);

  log(`creating answer`);
  const answer = await pc.createAnswer();
  await pc.setLocalDescription(answer);

  log(`adding answer to session`)
  await setAnswerOnSession({
    sdp: answer.sdp,
    type: answer.type,
  });

  // send local ICE candidates to signaling server
  pc.addEventListener("icecandidate", async (event) => {
    if (event.candidate) {
      log(`sending new local ICE candidate: ${event.candidate.toJSON().candidate}`);
      await addAnswerCandidate(event.candidate);
    }
  });

  // listen for new remote ICE candidates
  pollRemoteCandidatesHandle = setInterval(pollRemoteCandidates, 2000);
};

var addAnswerCandidate = async (candidate) => {
  await fetch(
    `${signalingServerURL}/addAnswerCandidate?id=` + sessionInput.value,
    {
      method: "POST",
      body: JSON.stringify(candidate),
    },
  );
};

let pollRemoteCandidates = async () => {
  let session = await getSession();

  // this is a bit hacky - we should only add newly seen remote candidates
  if (session && session.offer_candidates.length > 0) {
    clearInterval(pollRemoteCandidatesHandle);

    for (const c of session.offer_candidates) {
      const candidate = new RTCIceCandidate(c);
      log(`adding remote ICE candidate: ${candidate.toJSON().candidate}`);
      await pc.addIceCandidate(c);
    }
  }
};

let getSession = async () => {
  return (
    await fetch(`${signalingServerURL}/getSession?id=` + sessionInput.value)
  ).json();
};

let setAnswerOnSession = async (answer) => {
  return await fetch(
    `${signalingServerURL}/setAnswerOnSession?id=` + sessionInput.value,
    {
      method: "POST",
      body: JSON.stringify(answer),
    },
  );
};

let logsBlock = document.getElementById("logs");
let log = (msg) => {
  logsBlock.innerHTML += `${msg}<br>`;
};
