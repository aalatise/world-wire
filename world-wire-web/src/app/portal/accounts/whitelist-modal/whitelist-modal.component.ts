import { Component, OnInit, Inject } from '@angular/core';
import { BaseModal, ListItem, NotificationService } from 'carbon-components-angular';
import { AccountService } from '../../shared/services/account.service';
import { SessionService } from '../../../shared/services/session.service';
import { WhitelistRequest } from '../../shared/models/whitelist-request.interface';
import { UtilsService } from '../../../shared/utils/utils';
import { AuthService } from '../../../shared/services/auth.service';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { WorldWireError } from '../../../shared/models/error.interface';
import { PortalApiService } from '../../../shared/services/portal-api.service';

@Component({
  selector: 'app-whitelist-modal',
  templateUrl: './whitelist-modal.component.html',
  styleUrls: ['./whitelist-modal.component.scss']
})
export class WhitelistModalComponent extends BaseModal implements OnInit {

  requestedParticipant: string;

  // list of all active participants on the network
  participantOptions: ListItem[] = [];

  // id reference to the notification element in the UI to display errors
  notificationRef = '#notification';

  loadingRequest = false;

  participantsLoaded = false;

  constructor(
    private sessionService: SessionService,
    public accountService: AccountService,
    private authService: AuthService,
    private portalApiService: PortalApiService,
    private http: HttpClient,
    private notificationService: NotificationService,
    @Inject('MODAL_DATA') public data: any,
  ) {
    // MUST do super() to extend BaseModal
    super();
  }

  async ngOnInit() {

    const requests = [];

    // get list of all active participants on the network if not already retrieved
    if (!this.accountService.allParticipants) {
      requests.push(this.accountService.getAllParticipants());
    }

    // get list of whitelisted participants on the network if not already retrieved
    if (!this.accountService.whitelistedParticipants) {
      requests.push(this.accountService.getWhitelistedParticipants());
    }

    // make dynamic initial requests
    await Promise.all(requests);

    // populate dropdown options
    for (const participant of this.accountService.allParticipants) {

      // exclude current participant
      if (participant.id !== this.sessionService.currentNode.participantId) {

        const option: ListItem = {
          content: participant.id,
          value: participant.id,
          selected: false
        };

        // disable option if already in whitelist
        option.disabled = this.accountService.whitelistedParticipants && this.accountService.whitelistedParticipants.includes(participant.id);

        this.participantOptions.push(option);
      }
    }

    this.participantsLoaded = true;
  }

  /**
   * Creates new whitelist request
   * to be approved by a participant admin
   *
   * @returns
   * @memberof WhitelistModalComponent
   */
  async submitForm(): Promise<void> {
    if (!this.requestedParticipant) {
      // no participant was selected
      this.notificationService.showNotification({
        type: 'error',
        title: 'No Participant Selected',
        message: 'A participant must be selected to create a new whitelist request.',
        target: this.notificationRef
      });

      return;
    }

    try {

      this.loadingRequest = true;

      // maker action for whitelist request
      const request = `${this.accountService.globalRoot}/whitelist/v1/client/participants/whitelist`;

      let h: HttpHeaders = await this.authService.getFirebaseIdToken(this.sessionService.institution.info.institutionId, this.sessionService.currentNode.participantId);

      h = this.authService.addMakerCheckerHeaders(h, 'request');

      const options = {
        headers: h
      };

      const body = {
        participant_id: this.requestedParticipant
      };

      const response: WorldWireError = await this.http.post(
        request,
        body,
        options
      ).toPromise() as WorldWireError;

      if (response) {
        const requestBody: WhitelistRequest = {
          whitelisterId: this.sessionService.currentNode.participantId,
          whitelistedId: this.requestedParticipant,
          approvalIds: [response.msg],
        };

        this.portalApiService.createWhiteListRequest(requestBody).subscribe( () => {
            this.loadingRequest = false;
            this.closeModal();
          }, () => {
            this.loadingRequest = false;

            // show error message if unexpected error
            this.notificationService.showNotification({
              type: 'error',
              title: 'Unexpected Error',
              message: 'Unexpected error found when creating this trust request.',
              target: this.notificationRef
            });
          });
      } else {
        this.loadingRequest = false;

        // show error message if unexpected error
        this.notificationService.showNotification({
          type: 'error',
          title: 'Unexpected Error',
          message: 'Unexpected error found when creating this trust request.',
          target: this.notificationRef
        });
      }
    } catch (err) {

      this.loadingRequest = false;

      // show error message if unexpected error
      this.notificationService.showNotification({
        type: 'error',
        title: 'Unexpected Error',
        message: 'Unexpected error found when creating this trust request.',
        target: this.notificationRef
      });
    }
  }
}
