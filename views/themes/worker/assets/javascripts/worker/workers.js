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
  var CLASS_WORKER_CONTAINER = '.qor-worker-form';
  var CLASS_WORKER_LIST = '.qor-worker-form-list';
  var CLASS_WORKER_PROGRESS= '.qor-worker--progress';
  var CLASS_WORKER_SHOW= '.qor-worker-form--show';
  var CLASS_BUTTON_BACK= '.qor-button--back';
  var CLASS_TABLE= '.qor-js-table';
  var CLASS_SELECT= '.is-selected';

  function QorWorker(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorWorker.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorWorker.prototype = {
    constructor: QorWorker,

    init: function () {
      this.bind();
      this.formOpened = false;
      if ($(CLASS_WORKER_PROGRESS).size()) {
        $.fn.qorSliderAfterShow.updateWorkerProgress();
      }

      if ($(CLASS_WORKER_SHOW).size()) {
        this.formOpened = true;
      }

    },

    bind: function () {
      this.$element
        .on(EVENT_CLICK, CLASS_WORKER_ERRORS, $.proxy(this.showError, this))
        .on(EVENT_CLICK, CLASS_NEW_WORKER, $.proxy(this.showForm, this))
        .on(EVENT_CLICK, CLASS_BUTTON_BACK, $.proxy(this.hideForm, this));
    },

    unbind: function () {
      this.$element
        .off(EVENT_CLICK, CLASS_WORKER_ERRORS, this.showError, this)
        .off(EVENT_CLICK, CLASS_NEW_WORKER, this.showForm, this)
        .off(EVENT_CLICK, CLASS_BUTTON_BACK, this.hideForm, this);
    },

    showError: function (e) {
      e.preventDefault();

      var $workerErrorModal = $(QorWorker.POPOVERTEMPLATE).appendTo('body');
      var url = $('tr.is-selected .qor-button--edit').attr('href');
      $workerErrorModal.qorModal('show');

      $.ajax({
        url: url,
        method: 'GET',
        dataType: 'html',
        processData: false,
        contentType: false
      }).done(function (html) {
        var $content = $(html).find('.qor-form-container');
        var $errorTable = $content.find('.workers-error-output');
        if ($errorTable){
          $errorTable.appendTo($workerErrorModal.find('#qor-worker-errors'));
        }
      });

    },

    hideForm: function (e) {

      e.preventDefault();

      var $parent = this.$element;
      var $lists = $parent.find(CLASS_WORKER_CONTAINER).find('>li');

      $lists.show().removeClass('current').find('form').addClass('hidden');
      $(CLASS_BUTTON_BACK).addClass('hidden');
      $(CLASS_WORKER_LIST).show();

      this.formOpened = false;

      window.onbeforeunload = null;
      $.fn.qorSlideoutBeforeHide = null;

    },

    showForm: function (e) {
      var $target = $(e.target);
      e.preventDefault();

      if (this.formOpened) {
        return;
      }

      var $targetList = $target.closest('li');
      var $parent = $target.closest(CLASS_WORKER_CONTAINER);
      var $parentList = $target.closest(CLASS_WORKER_LIST);
      var $lists = $parent.find('>li');

      $lists.hide().removeClass('current');

      $targetList.addClass('current').show();
      $(CLASS_BUTTON_BACK).removeClass('hidden');
      $targetList.find(CLASS_WORKER_LIST).hide();

      $parentList.show().find('form').removeClass('hidden');
      this.formOpened = true;
    },

    destroy: function () {
      this.unbind();
      QorWorker.getWorkerProgressIntervId && window.clearInterval(QorWorker.getWorkerProgressIntervId);
      $.fn.qorSliderAfterShow.updateWorkerProgress = null;
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

  QorWorker.updateTableStatus = function (status) {
    var $selectedItem = $(CLASS_TABLE).find(CLASS_SELECT);
    var statusName = $(CLASS_WORKER_PROGRESS).data().statusName;

    $selectedItem.find('td[data-heading="' + statusName + '"]').find('.qor-table__content').html(status);

  };

  QorWorker.isScrollToBottom = function (element) {
    return element.clientHeight + element.scrollTop === element.scrollHeight;
  };

  QorWorker.updateWorkerProgress = function (url) {
    var progressURL = url;
    var $logContainer = $('.workers-log-output');
    var $progressValue = $('.qor-worker--progress-value');
    var $progressStatusStatus = $('.qor-worker--progress-status');
    var $progress = $(CLASS_WORKER_PROGRESS);
    var $selectTR = $(CLASS_TABLE).find(CLASS_SELECT);
    var status = ['killed','exception','cancelled','scheduled'];

    if ($progress.size()) {
      var progressData = $progress.data();
    }

    if ($selectTR.size() && progressData && progressData.statusName) {
      var orignialStatus = $selectTR.find('td[data-heading="' + progressData.statusName + '"]').find('.qor-table__content').html();
    }

    if (!$progress.size() || !$progress.size() || status.indexOf(progressData.status) != -1) {
      window.clearInterval(QorWorker.getWorkerProgressIntervId);
      return;
    }

    if (progressData.progress >= 100){
      window.clearInterval(QorWorker.getWorkerProgressIntervId);
      document.querySelector('#qor-worker--progress').MaterialProgress.setProgress(100);
      QorWorker.updateTableStatus(progressData.status);
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
      var $html = $(html);
      var contentData = $html.find(CLASS_WORKER_PROGRESS).data();
      var currentStatus = contentData.progress;
      var progressStatusStatus = contentData.status;
      $progressValue.html(currentStatus);
      $progressStatusStatus.html(progressStatusStatus);

      // set status progress
      document.querySelector('#qor-worker--progress').MaterialProgress.setProgress(currentStatus);

      // update process log
      var oldLog = $.trim($logContainer.html());
      var newLog = $.trim($html.find('.workers-log-output').html());
      var newLogHtml;

      if (newLog != oldLog){
        newLogHtml = newLog.replace(oldLog, '');

        if (QorWorker.isScrollToBottom($logContainer[0])){
          $logContainer.append(newLogHtml).scrollTop($logContainer[0].scrollHeight);
        } else {
          $logContainer.append(newLogHtml);
        }

      }

      if (orignialStatus != progressStatusStatus) {
        QorWorker.updateTableStatus(progressStatusStatus);
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
