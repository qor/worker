$(function () {

  'use strict';

  $(document).on('click.qor.worker.trigger', '.qor-sub-nav-trigger', function (e) {
    var terget = $(e.currentTarget);
    var targetParent = terget.parent();

    targetParent.addClass('current');
    $('.qor-worker-form-list').not('current').find('form').addClass('hidden');
    terget.next('form').toggleClass('hidden');
  });

});
