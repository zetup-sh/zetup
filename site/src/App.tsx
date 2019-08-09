import React, {useState} from 'react';
import './App.scss';


const OS = getOS()
const cmds = {
  "curl": "curl -L -O https://raw.github.com/zetup-sh/zetup/build/\"\n",

  "wget": "sh -c \"$(wget https://raw.github.com/zeutp-sh/zetup/master/tools/install.sh -O -)\"",
  "powershell": "Coming soon to a terminal near you!",
}


const App: React.FC = () => {
  const [curCmd, setCurCmd] = useState(OS === "Windows" ? "powershell" : "curl")
  const copyCmd = () => {
    copyTextToClipboard(cmds[curCmd])
  }

  return (
    <div>
    <div id="fog" />
    <h1>Z</h1>
    <div className="install-menu">
    <h3>Installation</h3>
    <div className="tab">
      {Object.entries(cmds).map(([label]) => (
        <button
          key={label}
          className={"tablinks"+ (curCmd === label ? " active" : "")}
          onClick={()=> setCurCmd(label)}>
        {label}
       </button>
      ))}
    </div>
    <div className="tabcontent">
      <code id="install-cmd">{cmds[curCmd]}</code>
      {curCmd !== "powershell" && (
        <button className="btn" onClick={copyCmd}><i className="fa fa-copy"></i></button>
      )}
    </div>
    </div>
    </div>
  );
}

export default App;

function getOS() {
  var userAgent = window.navigator.userAgent,
      platform = window.navigator.platform,
      macosPlatforms = ['Macintosh', 'MacIntel', 'MacPPC', 'Mac68K'],
      windowsPlatforms = ['Win32', 'Win64', 'Windows', 'WinCE'],
      iosPlatforms = ['iPhone', 'iPad', 'iPod']
  let os = ""

  if (macosPlatforms.indexOf(platform) !== -1) {
    os = 'Mac OS';
  } else if (iosPlatforms.indexOf(platform) !== -1) {
    os = 'iOS';
  } else if (windowsPlatforms.indexOf(platform) !== -1) {
    os = 'Windows';
  } else if (/Android/.test(userAgent)) {
    os = 'Android';
  } else if (!os && /Linux/.test(platform)) {
    os = 'Linux';
  }

  return os;
}


// https://stackoverflow.com/a/30810322
function fallbackCopyTextToClipboard(text) {
  var textArea = document.createElement("textarea");
  textArea.value = text;
  document.body.appendChild(textArea);
  textArea.focus();
  textArea.select();

  try {
    var successful = document.execCommand('copy');
    var msg = successful ? 'successful' : 'unsuccessful';
    console.log('Fallback: Copying text command was ' + msg);
  } catch (err) {
    console.error('Fallback: Oops, unable to copy', err);
  }

  document.body.removeChild(textArea);
}

function copyTextToClipboard(text) {
  if (!navigator.clipboard) {
    fallbackCopyTextToClipboard(text);
    return;
  }
  navigator.clipboard.writeText(text).then(function() {
    console.log('Async: Copying to clipboard was successful!');
  }, function(err) {
    console.error('Async: Could not copy text: ', err);
  });
}