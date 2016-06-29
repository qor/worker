(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define(['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var NAMESPACE = 'qor.worker';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var CLASS_NEW_WORKER = '.qor-worker--new';
  var CLASS_WORKER_ERRORS = '.qor-worker--show-errors';
  var CLASS_WORKER_LIST = '.qor-worker-form-list';
  var CLASS_WORKER_PROGRESS= '.qor-worker--progress';

  function QorWorker(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorWorker.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorWorker.prototype = {
    constructor: QorWorker,

    init: function () {
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.click, this));
    },

    click: function (e) {
      var $target = $(e.target);
      e.stopPropagation();

      if ($target.is(CLASS_WORKER_ERRORS)){
        var $workerErrorModal = $(QorWorker.POPOVERTEMPLATE).appendTo('body');
        var url = $('tr.is-selected .qor-button--edit').attr('href');
        $workerErrorModal.qorModal('show');

        $.ajax({
          url: url
        }).done(function (html) {
          var $content = $(html).find('.qor-form-container');
          var $errorTable = $content.find('.workers-error-output');
          if ($errorTable){
            $errorTable.appendTo($workerErrorModal.find('#qor-worker-errors'));
          }
        });
      }

      if ($target.is(CLASS_NEW_WORKER)){
        var $targetParent = $target.closest(CLASS_WORKER_LIST);

        $(CLASS_WORKER_LIST).removeClass('current');
        $targetParent.addClass('current');

        var $list = $(CLASS_WORKER_LIST).not('.current');

        $list.find('form').addClass('hidden');
        $list.find(CLASS_NEW_WORKER).removeClass('open');

        $target.next('form').toggleClass('hidden');
        $target.toggleClass('open');
      }
    }
  };

  QorWorker.DEFAULTS = {};
  QorWorker.POPOVERTEMPLATE = (
     '<div class="qor-modal fade qor-modal--worker-errors" tabindex="-1" role="dialog" aria-hidden="true">' +
      '<div class="mdl-card mdl-shadow--2dp" role="document">' +
        '<div class="mdl-card__title">' +
          '<h2 class="mdl-card__title-text">Process Errors</h2>' +
        '</div>' +
        '<div class="mdl-card__supporting-text" id="qor-worker-errors"></div>' +
        '<div class="mdl-card__actions">' +
          '<a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect" data-dismiss="modal">close</a>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  QorWorker.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {

        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorWorker(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $.fn.qorSliderAfterShow.updateWorkerProgress = function (url) {
    QorWorker.getWorkerProgressIntervId = window.setInterval(QorWorker.updateWorkerProgress, 1000, url);
  };

  QorWorker.isScrollToBottom = function (element) {
    return element.clientHeight + element.scrollTop === element.scrollHeight;
  };

  QorWorker.updateWorkerProgress = function (url) {
    var progressURL = url;
    var $logContainer = $('.workers-log-output');
    var $progressValue = $('.qor-worker--progress-value');
    var $progressStatusStatus = $('.qor-worker--progress-status');
    var workerProgress = document.querySelector('#qor-worker--progress');

    console.log($(CLASS_WORKER_PROGRESS));
    console.log('progress:' + $(CLASS_WORKER_PROGRESS).data('worker-progress'));

    if (!$(CLASS_WORKER_PROGRESS).size()) {
      window.clearInterval(QorWorker.getWorkerProgressIntervId);
    }

    if ($(CLASS_WORKER_PROGRESS).data('worker-progress') >= 100){
      window.clearInterval(QorWorker.getWorkerProgressIntervId);
      workerProgress.MaterialProgress.setProgress(100);
      $('.qor-workers-abort').addClass('hidden');
      $('.qor-workers-rerun').removeClass('hidden');
      return;
    }

    $.ajax({
      url: progressURL,
      method: 'GET',
      dataType: 'html',
      processData: false,
      contentType: false
    }).done(function (html) {
      var $content = $(html).find(CLASS_WORKER_PROGRESS);
      var currentStatus = $content.data('worker-progress');
      var progressStatusStatus = $content.data('worker-status');

      $progressValue.html(currentStatus);
      $progressStatusStatus.html(progressStatusStatus);

      // set status progress
      if (workerProgress && workerProgress.MaterialProgress){
        workerProgress.MaterialProgress.setProgress(currentStatus);
      }

      // update process log
      var oldLog = $.trim($logContainer.html());
      var newLog = $.trim($content.find('.workers-log-output').html());
      var newLogHtml;

      if (newLog != oldLog){
        newLogHtml = newLog.replace(oldLog, '');

        if (QorWorker.isScrollToBottom($logContainer[0])){
          $logContainer.append(newLogHtml).scrollTop($logContainer[0].scrollHeight);
        } else {
          $logContainer.append(newLogHtml);
        }

      }

      if (currentStatus >= 100){
        window.clearInterval(QorWorker.getWorkerProgressIntervId);
        $('.qor-workers-abort').addClass('hidden');
        $('.qor-workers-rerun').removeClass('hidden');
      }

    });
  };


  $(function () {
    var selector = '[data-toggle="qor.workers"]';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorWorker.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorWorker.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorWorker;

});
