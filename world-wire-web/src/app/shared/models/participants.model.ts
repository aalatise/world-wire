import { IInstitution } from './participant.interface';
import { Injectable, isDevMode, OnDestroy } from '@angular/core';
import { Subscription } from 'rxjs';
import { FI_TYPES, ROLES, ENVIRONMENT, STATUS } from '../constants/general.constants';
import * as _ from 'lodash';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../environments/environment';
import { PortalApiService } from '../services/portal-api.service';

// export interface IInstitutionAll {
//     // array of institutions
//     arr: IInstitution[];
//     // used to lookup by institution by slug (both should be unique)
//     ids: { [slug: string]: /*institutionId*/ string }[];
// }

@Injectable()
export class ParticipantsModel implements OnDestroy {

    /**
     * Primarily used to get details about a participant
     * including a list of associated users with permissions.
     * Security: NOT a secure node for evaluating permissions
     * @memberof ParticipantModel
     */
    route = 'participants/{institutionId}'; // firebase ref

    model: IInstitution;

    types = FI_TYPES;
    roles = ROLES;
    status = STATUS;
    searchedParticipants: IInstitution[];

    private subscriptions: Subscription[];

    constructor(
        private http: HttpClient,
        private portalApiService: PortalApiService,
        // Note: Do not use ngZone here (only use in component view) so that this class can be easily tested.
        // private ngZone: NgZone
    ) {

        this.model = {
            info: {
                institutionId: '',
                name: '',
                geo_lat: '',
                geo_lon: '',
                country: '',
                address1: '',
                address2: '',
                city: '',
                state: '',
                zip: '',
                logo_url: '',
                site_url: '',
                kind: 'Money Transfer Operator',
                slug: '',
                status: 'active'
            }
        };

        this.subscriptions = [];

    }

    ngOnDestroy() {
        this.subscriptions.forEach(sub => sub.unsubscribe());
    }

    /**
     * Sets the slug to a kebab case formated string
     * @param text
     * @memberof ParticipantsModel
     */
    genSlug(text: string) {
        // with type info
        this.model.info.slug = _.kebabCase(text);
    }

    /**
    * Returns all participants listed in firebase database
    * (not participant registry). Firebase database is
    * kept separately in sync with PR so that if anything
    * malicious were to be done to the firebase database
    * of participants it would not affect the WW network
    *
    * @returns {Promise<{ [id: string]: IInstitution }>}
    * @memberof ParticipantsModel
    */
    get(institutionId: string): Promise<IInstitution> {
        return this.portalApiService.getInstitutionP(institutionId);
    }

    /**
     * Returns all participants listed in firebase database
     * (not participant registry). Firebase database is
     * kept separately in sync with PR so that if anything
     * malicious were to be done to the firebase database
     * of participants it would not affect the WW network
     *
     * @returns {Promise<{ [id: string]: IInstitution }>}
     * @memberof ParticipantsModel
     */
    allPromise(): Promise<IInstitution[]> {

        return new Promise((resolve, reject) => {

            const participantObs = this.portalApiService.getAllInstitution();

            const participantSub = participantObs.subscribe((data: IInstitution[]) => {
                // sort by 'name' field
                const participants = _.sortBy(data,
                    (o: IInstitution) => {
                        // sort case insensitive
                        return o.info.name.toLowerCase();
                    }
                );

                resolve(participants);

            }, err => {
                if (isDevMode()) {
                    console.log(err);
                }
                reject('Unable to get all participants');
            });

            this.subscriptions.push(participantSub);

        });

    }

    // /**
    //  * Returns all participants listed in firebase database
    //  * (not participant registry). Firebase database is
    //  * kept separately in sync with PR so that if anything
    //  * malicious were to be done to the firebase database
    //  * of participants it would not affect the WW network
    //  *
    //  * @returns {Observable<{ [id: string]: IInstitution }>}
    //  * @memberof ParticipantsModel
    //  */
    // allObservable(): Observable<IInstitution[]> {

    //     const source = new Observable((observer: Observer<IInstitution[]>) => {

    //         // watch changes to values affecting firebase ref
    //         this.allRef.on('value',
    //             (participants: DatabaseSnapshot<{ [id: string]: IInstitution }>) => {

    //                 // sort by 'name' field
    //                 const data = _.sortBy(
    //                     // return in array format
    //                     _.toArray(participants.val()),
    //                     (o: IInstitution) => {
    //                         // sort case insensitive
    //                         return o.info.name.toLowerCase();
    //                     }
    //                 );

    //                 // since calling from external source
    //                 // need to put result into angular zone
    //                 // this.ngZone.run(() => {
    //                 // update observer value
    //                 observer.next(
    //                     data
    //                 );

    //                 // });

    //             }, (error: any) => {
    //                 console.log(error);

    //             }
    //         );

    //     });

    //     return source;

    // }

    // /**
    //  * This detaches a callback so that callback listeners
    //  * are appropriately removed when no longer used.
    //  * This should be used with 'ngOnDestroy' since leaving a
    //  * angular component usually means that the data associated
    //  * with this callback is no longer needed
    //  *
    //  * @memberof ParticipantsModel
    //  */
    // unsubscribeObservable() {

    //     // console.log('off');

    //     // stop listener on
    //     // https://firebase.google.com/docs/reference/js/firebase.database.Reference#off
    //     this.allRef.off();
    // }

    /**
    * Creates a new participant
    *
    * @memberof ParticipantModel
    */
    create(): Promise<boolean> {

        return new Promise(async (resolve, reject) => {
            try {
                await this.portalApiService.createInstitutionP(this.model);
                resolve(true);
            } catch (err) {
                if (isDevMode()) {
                    console.log(err);
                }
                reject(false);
            }
        });

    }

    /**
     * Updates an existing participant's info
     *
     * @returns
     * @memberof ParticipantModel
     */
    update(): Promise<boolean> {
      return new Promise(async (resolve, reject) => {
        if (this.model.info.institutionId) {
          try {
            await this.portalApiService.updateInstitutionP(this.model.info.institutionId, this.model);
            resolve(true);
          } catch (err) {
            if (isDevMode()) {
              console.log(err);
            }
            reject(false);
          }
        } else {
          console.log('No institutionId provided.');
          reject(false);
        }
      });
    }

    delete(institutionId: string) {
        return new Promise(async (resolve, reject) => {

            if (institutionId) {

                try {
                    await this.portalApiService.deleteInstitutionP(institutionId);
                    resolve();
                } catch (err) {
                    if (isDevMode()) {
                        console.log(err);
                    }
                    reject();
                }

            } else {
                console.log('No institutionId provided.');
                reject();
            }

        });
    }

}
