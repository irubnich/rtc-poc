const config = {
    iceServers: [
      {
        urls: ['stun:stun1.l.google.com:19302', 'stun:stun2.l.google.com:19302'],
      },
    ],
    iceCandidatePoolSize: 10,
  }
  
  const pc = new RTCPeerConnection(config);
  pc.addTransceiver("audio")
  
  pc.addEventListener('connectionstatechange', event => {
    if (pc.connectionState === 'connected') {
      console.log("CONNECTED")
    }
  })
  
  let localChannel;
  
  pc.addEventListener("datachannel", event => {
    event.channel.addEventListener("message", event => {
      console.log("got message", event.data)
    })
  })
  
  let c = pc.createDataChannel("send")
  c.addEventListener("open", event => {
    localChannel = event.currentTarget;
    c.send("hi")
  })
  
  let pollAnswerHandle;
  let pollAnswerCandidatesHandle;
  let pollOfferCandidatesHandle;
  
  let sessionInput = document.getElementById("session");
  
  let sendBtn = document.getElementById("send-btn");
  sendBtn.onclick = event => {
    localChannel.send("hi")
  }
  
  let clientBtn = document.getElementById("client-btn");
  clientBtn.onclick = async () => {
    pc.addEventListener('icecandidate', event => {
      if (event.candidate) {
        addOfferCandidate(event.candidate);
      }
    })
  
    const offerDescription = await pc.createOffer();
    await pc.setLocalDescription(offerDescription);
  
    const offer = {
      sdp: offerDescription.sdp,
      type: offerDescription.type
    };
  
    await createSession(offer);
  
    pollAnswerHandle = setInterval(pollAnswer, 2000);
    pollAnswerCandidatesHandle = setInterval(pollAnswerCandidates, 2000);
  };
  
  let serverBtn = document.getElementById("server-btn");
  serverBtn.onclick = async () => {
    pc.addEventListener('icecandidate', event => {
      if (event.candidate) {
        addAnswerCandidate(event.candidate);
      }
    })
  
    let session = await getSession();
    let offer = session.offer;
    await pc.setRemoteDescription(new RTCSessionDescription(offer));
  
    const ans = await pc.createAnswer();
    await pc.setLocalDescription(ans);
  
    const answer = {
      sdp: ans.sdp,
      type: ans.type
    };
  
    setAnswerOnSession(answer);
    pollOfferCandidatesHandle = setInterval(pollOfferCandidates, 2000);
  };
  
  var createSession = async (session) => {
    await fetch("http://signaling.jimbo.sh:8080/createSession?id=" + sessionInput.value, {
      method: "POST",
      body: JSON.stringify(session)
    })
  };
  
  var setAnswerOnSession = async (session) => {
    await fetch("http://signaling.jimbo.sh:8080/setAnswerOnSession?id=" + sessionInput.value, {
      method: "POST",
      body: JSON.stringify(session)
    })
  };
  
  var addOfferCandidate = async (candidate) => {
    await fetch("http://signaling.jimbo.sh:8080/addOfferCandidate?id=" + sessionInput.value, {
      method: "POST",
      body: JSON.stringify(candidate)
    })
  };
  
  var addAnswerCandidate = async (candidate) => {
    await fetch("http://signaling.jimbo.sh:8080/addAnswerCandidate?id=" + sessionInput.value, {
      method: "POST",
      body: JSON.stringify(candidate)
    })
  };
  
  let pollAnswer = async () => {
    let session = await getSession();
    if (session && session.answer) {
      clearInterval(pollAnswerHandle);
  
      const answerDesc = new RTCSessionDescription(session.answer);
      await pc.setRemoteDescription(answerDesc);
    }
  };
  
  let pollAnswerCandidates = async () => {
    let session = await getSession();
    if (session && session.answer_candidates.length > 0) {
      clearInterval(pollAnswerCandidatesHandle);
  
      const candidate = new RTCIceCandidate(session.answer_candidates[0]);
      await pc.addIceCandidate(candidate);
    }
  };
  
  let pollOfferCandidates = async () => {
    let session = await getSession();
    if (session && session.offer_candidates.length > 0) {
      clearInterval(pollOfferCandidatesHandle);
  
      const candidate = new RTCIceCandidate(session.offer_candidates[0]);
      await pc.addIceCandidate(candidate);
    }
  };
  
  let getOffers = async () => {
    return (await fetch("http://signaling.jimbo.sh:8080/getOffers?id=" + sessionInput.value)).json()
  }
  
  let getSession = async () => {
    return (await fetch("http://signaling.jimbo.sh:8080/getSession?id=" + sessionInput.value)).json()
  }
  