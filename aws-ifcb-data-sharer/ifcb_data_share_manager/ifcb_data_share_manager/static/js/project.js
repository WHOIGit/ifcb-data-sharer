/* Project specific Javascript goes here. */
var example2Left = document.getElementById('example2-left'),
	example2Right = document.getElementById('example2-right')


var leftSort = new Sortable(example2Left, {
    group: 'shared', // set both lists to same group
    multiDrag: true, // Enable multi-drag
	selectedClass: 'active', // The class applied to the selected items
	fallbackTolerance: 3, // So that we can select items on mobile
	animation: 150,
    filter: '.list-group-item-primary'
});

var rightSort = new Sortable(example2Right, {
    group: 'shared',
    multiDrag: true, // Enable multi-drag
	selectedClass: 'active', // The class applied to the selected items
	fallbackTolerance: 3, // So that we can select items on mobile
	animation: 150,
    filter: '.list-group-item-primary'
});

events = [
    'onChoose',
    'onStart',
    'onEnd',
    'onAdd',
    'onUpdate',
    'onSort',
    'onRemove',
    'onChange',
    'onUnchoose',
    'onSelect',
    'onDeselect'
  ].forEach(function (name) {
    leftSort.options[name] = function (evt) {
      console.log({
        'event': name,
        'this': this,
        'item': evt.item,
        'items': evt.items,
        'clones': evt.clones,
        'from': evt.from,
        'to': evt.to,
        'oldIndex': evt.oldIndex,
        'newIndex': evt.newIndex
      });
    };
  });