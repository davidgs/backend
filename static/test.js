const fb = document.getElementById(type.toLowerCase());
const butt = document.createElement('button');
butt.setAttribute('class', 'btn btn-success dropdown-toggle');
butt.setAttribute('type', 'button');
butt.setAttribute('data-bs-toggle', 'dropdown');
butt.setAttribute('aria-expanded', 'false');
butt.setAttribute('name', lastPart.toLowerCase());
butt.innerHTML = 'Select ' + lastPart;
const list = document.createElement('ul');
list.setAttribute('class', 'dropdown-menu');
list.setAttribute('aria-labelledby', 'dropdownMenuButton1');
list.setAttribute('id', lastPart[0].toUpperCase() + lastPart.slice(1).toLowerCase());

for (var x = 0; x < data.length; x++) {
  const li = document.createElement('li');
  li.setAttribute('id', lastPart.toLowerCase() + '-' + data[x]._id);
  const button = document.createElement('button');
  button.setAttribute('id', data[x]._id);
  button.setAttribute('value', lastPart.toLowerCase());
  button.setAttribute('onclick', 'selectedChoice(this.id, this.value)');
  button.setAttribute('class', 'dropdown-item');
  button.setAttribute('type', 'button');
  button.innerHTML = data[x].Name;
  li.appendChild(button);
  list.appendChild(li);
}
butt.appendChild(list);
fb.replaceChildren(butt);

document.getElementById(lastPart.toLowerCase()).innerHTML = selectVal;