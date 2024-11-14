//const signalingServerURL = "https://autodiscovery-signaling.app-builder-on-prem.net";
const signalingServerURL = "http://localhost:8080";

const config = {
  iceServers: [
    {
      urls: ["stun:stun1.l.google.com:19302", "stun:stun2.l.google.com:19302"],
    },
  ],
  iceCandidatePoolSize: 10,
};

const pc = new RTCPeerConnection(config);
pc.addTransceiver("audio");

let logsBlock = document.getElementById("logs");
let log = (msg) => {
  logsBlock.innerHTML += `${msg}<br>`;
};

pc.addEventListener("connectionstatechange", (event) => {
  log(`new connection state: ${pc.connectionState}`);
});

let localChannel;

let sendBtn = document.getElementById("send-btn");
sendBtn.onclick = (event) => {
  log(`sending a message over the data channel`);
  localChannel.send("hi");
};

pc.addEventListener("datachannel", (event) => {
  localChannel = event.channel;
  event.channel.addEventListener("message", (event) => {
    const decodedMsg = new TextDecoder("utf-8").decode(event.data);
    log(`got message on datachannel: ${decodedMsg}`);
  });

  localChannel.onopen = () => {
    sendBtn.disabled = false;
  };
});

let pollRemoteCandidatesHandle;

let sessionInput = document.getElementById("session");

let clientBtn = document.getElementById("client-btn");
clientBtn.onclick = async () => {
  clientBtn.disabled = true;

  log(`getting session`);
  let session = await getSession();

  log(`setting remote offer`);
  await pc.setRemoteDescription(session.offer);

  log(`creating answer`);
  const answer = await pc.createAnswer();
  await pc.setLocalDescription(answer);

  const answerSerialized = {
    sdp: answer.sdp,
    type: answer.type,
  };
  await setAnswerOnSession(answerSerialized);

  pc.addEventListener("icecandidate", async (event) => {
    if (event.candidate) {
      log(`got ICE candidate: ${event.candidate.toJSON().candidate}`);
      await addAnswerCandidate(event.candidate);
    }
  });

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
  if (session && session.offer_candidates.length > 0) {
    clearInterval(pollRemoteCandidatesHandle);

    for (const c of session.offer_candidates) {
      log(`got remote candidate!`);
      const candidate = new RTCIceCandidate(c);
      await pc.addIceCandidate(candidate);
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
