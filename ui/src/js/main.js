const input = document.querySelector("input[type=file]");
const preview = document.querySelector(".preview");

// input.style.opacity = 0;
input.addEventListener("change", updateImageDisplay);
function updateImageDisplay() {
  while (preview.firstChild) {
    preview.removeChild(preview.firstChild);
  }

  const curFiles = input.files;
  if (curFiles.length === 0) {
    const para = document.createElement("p");
    para.textContent = "No files currently selected for upload";
    preview.appendChild(para);
  } else {
    const list = document.createElement("ol");
    preview.appendChild(list);

    for (const file of curFiles) {
      const listItem = document.createElement("li");
      const para = document.createElement("p");
      const [name, size] = [document.createElement("span"), document.createElement("span")];

      name.classList.add("file-name");
      size.classList.add("file-size");
      para.style = "display: inline-block";

      name.textContent = file.name;

      if (validFileType(file)) {
        size.textContent = `, ${returnFileSize(file.size)}`;
        para.appendChild(name)
        para.appendChild(size)
        listItem.appendChild(para);
      } else {
        para.appendChild(name)
        name.textContent = `${file.name}: Not a valid file type. Update your selection.`;
        listItem.appendChild(para);
      }

      list.appendChild(listItem);
    }
  }
}

const fileTypes = [
  "font/ttf",
  "font/woff",
  "font/woff2",
];

function validFileType(file) {
  return fileTypes.includes(file.type);
}
function returnFileSize(number) {
  const [kb, mb] = [1 << 10, 1 << 20];

  if (number < kb) {
    return `${number} bytes`;
  } else if (number >= kb && number < mb) {
    return `${(number / kb).toFixed(1)} KB`;
  }
  return `${(number / mb).toFixed(1)} MB`;
}
// const button = document.querySelector("form button");
// button.addEventListener("click", (e) => {
//   e.preventDefault();
//   const para = document.createElement("p");
//   para.append("Image uploaded!");
//   preview.replaceChildren(para);
// });
